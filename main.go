package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"github.com/BHunter2889/da-fish-alexa/services"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"log"
	"net/http"
	"strings"
)

var (
	cfg           *BugCasterConfig
	defaultUserIp = "127.0.0.1"

	deviceLocService  services.DeviceService
	GeocodeService    services.GeocodeService
	ForecasterService services.ForecasterService

	tombLoc *DeviceLocator
)

func IntentDispatcher(ctx context.Context, request alexa.Request) alexa.Response {
	log.Print("Intent Dispatcher")
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "TodaysFishRatingIntent":
		log.Print("INTENT_DISPATCH: TodaysFishRatingIntent")
		response = HandleTodaysFishRatingIntent(ctx, request)
	case alexa.HelpIntent:
		log.Print("INTENT_DISPATCH: HelpIntent")
		response = HandleHelpIntent(ctx, request)
	case "AboutIntent":
		log.Print("INTENT_DISPATCH: AboutIntent")
		response = HandleAboutIntent(ctx, request)
	default:
		log.Print("INTENT_DISPATCH: Default Response")
		// Kill Running processes, record the error, move on to returning the default response.
		killAllTombs(ctx)
		response = HandleAboutIntent(ctx, request)
	}
	return response
}

func HandleTodaysFishRatingIntent(ctx context.Context, request alexa.Request) (response alexa.Response) {

	err := xray.Capture(ctx, "TodaysFishRatingIntent", func(ctx1 context.Context) error {

		log.Print("Todays Fish Rating Intent")

		// Handle Any Panics with a Default Error Response
		defer func() {
			if r := recover(); r != nil {
				log.Print("Benzos delivered in TodaysFishRatingIntent handler.")

				var xErr error
				switch tErr := r.(type) {
				case string:
					xErr = errors.New(tErr)
				case error:
					xErr = tErr
				default:
					log.Print(r)
					xErr = errors.New("unspecified error")
				}

				addAndHandleXRayRecordingError(ctx, xErr)
				response = alexa.NewDefaultErrorResponse()
			}
		}()

		// Get Location registered to user device
		var locResp *alexa.DeviceLocationResponse
		locErr := xray.Capture(ctx1, "TodaysFishRatingIntent.deviceLocation", func(ctx2 context.Context) error {

			deviceId := request.Context.System.Device.DeviceID
			apiAccessToken := request.Context.System.APIAccessToken
			apiEndpoint := request.Context.System.APIEndpoint

			deviceLocService = services.DeviceService{
				URL:      apiEndpoint,
				Id:       deviceId,
				Token:    apiAccessToken,
				Endpoint: cfg.AlexaLocEndpoint,
			}

			tombLoc := NewDeviceLocator(ctx, &deviceLocService)
			var err error

			// Listen to select either a Successful Response, or Error, whichever comes first.
			select {
			case err = <-tombLoc.ErCh:
				if err != nil {

					// Expected case when Location Permissions have not yet been enabled.
					if strings.Contains(err.Error(), "403") {
						log.Print("REQUESTING LOCATION PERMISSIONS")

						// Kill any processes still running, Record the Error, return new Permissions Request Response
						killAllTombs(ctx)
						response = alexa.NewPermissionsRequestResponse()
						return nil
					} else { // Trouble Communicating with Alexa Devices Api
						log.Print("Unexpected Device Location Retrieval Error")

						// Kill Processes, Record Error, return fallback response.
						killAllTombs(ctx)
						panic("Unexpected Device Location Retrieval Error")
					}
				}
			case locResp = <-tombLoc.Ch:
				addAndHandleXRayRecordingError(ctx2, xray.AddMetadata(ctx2, "device-location", locResp))
				log.Print(locResp)
			}
			return nil
		})

		addAndHandleXRayRecordingError(ctx1, locErr)

		// Wait for kMS decryption service calls to finish if they haven't before proceeding.
		KMSDecryptionWaiter()

		// Get Geocode coordinates from retrieved location
		var geoPoint services.GeoPoint
		geoErr := xray.Capture(ctx1, "TodaysFishRatingIntent.geocodeLocation", func(ctx2 context.Context) error {

			GeocodeService = services.GeocodeService{
				URL:           cfg.GeoUrl,
				UsrIp:         defaultUserIp,
				CountryRegion: locResp.CountryCode,
				PostalCode:    locResp.PostalCode,
				Key:           cfg.GeoKey,
				Client:        http.Client{},
			}
			log.Print(cfg.GeoKey)
			log.Print(cfg.FishRatingUrl)
			log.Print("Geocoding Location... ")

			var err error
			geoPoint, err = GeocodeService.GetGeoPoint(ctx)
			if err != nil {
				log.Print("trouble getting information on the area ")
				log.Print(err)
				panic("trouble getting information on the area ")
			}
			log.Printf("Geo: {lat: %f, lon: %f}", geoPoint.Coordinates[0], geoPoint.Coordinates[1])

			// Add Geocoded Coordinates to the Encrypted X-Ray Traces
			addAndHandleXRayRecordingError(ctx2, xray.AddMetadata(ctx2, "geocoded-location", geoPoint))
			return nil
		})

		addAndHandleXRayRecordingError(ctx1, geoErr)

		//Get Fishing Forecast using coordinates
		var fr []services.Hour
		fishErr := xray.Capture(ctx1, "TodaysFishRatingIntent.fishingForecast", func(ctx2 context.Context) error {

			ForecasterService = services.ForecasterService{
				URL:    cfg.FishRatingUrl,
				Lat:    geoPoint.Coordinates[0],
				Lon:    geoPoint.Coordinates[1],
				Client: http.Client{},
			}
			log.Print("Getting Forecast... ")
			var err error
			fr, err = ForecasterService.GetCurrentFishingRating(ctx)
			if err != nil {
				log.Print(err)
				log.Print("Trouble Getting Fishing Forecast ")
				panic("Trouble Getting Fishing Forecast ")
			}

			addAndHandleXRayRecordingError(ctx, xray.AddMetadata(ctx2, "forecast-hours", fr))
			return nil
		})

		addAndHandleXRayRecordingError(ctx1, fishErr)
		var (
			t string
			r int
			w float64
		)
		if fr[0].Rating > fr[1].Rating {
			t, r, w = "right now", fr[0].Rating, fr[0].WindSpeed
		} else {
			t, r, w = "two hours from now", fr[1].Rating, fr[1].WindSpeed
		}

		// TODO Consider moving these response builds to a separate file.
		var fcstBuilder alexa.SSMLBuilder
		if r < 2 {
			fcstBuilder.Say("It looks like the best time to go fishing over the next couple of hours is, ")
			fcstBuilder.Pause("150")
			fcstBuilder.Say("well, ")
			fcstBuilder.Pause("150")
			fcstBuilder.Say("probably some other time.")
			fcstBuilder.Pause("500")
			fcstBuilder.Say(fmt.Sprintf(" The top rating is well below average and the wind speed is %.1f miles per hour.", w))
		} else if r >= 2 && r <= 3 {
			fcstBuilder.Say(fmt.Sprintf("It looks like a decent or possibly better time to go fishing %s.", t))
			fcstBuilder.Say(" The forecast rating is on the plus side,")
			fcstBuilder.Pause("100")
			fcstBuilder.Say("So you should have at least an average time fishing.")
			fcstBuilder.Pause("500")
			fcstBuilder.Say(fmt.Sprintf("The wind speed is listed at %.1f miles per hour.", w))
		} else {
			fcstBuilder.Say("The fish appear to be biting!")
			fcstBuilder.Pause("750")
			fcstBuilder.Say("Over the next couple hours the fishing looks great!")
			fcstBuilder.Pause("500")
			fcstBuilder.Say(fmt.Sprintf(
				" The best time to go appears to be %s with a forecast rating well over the norm. ", t))
			fcstBuilder.Pause("500")
			fcstBuilder.Say(fmt.Sprintf("The wind speed is listed at %.1f miles per hour.", w))
		}

		response = alexa.NewSSMLResponse("Today's Fishing Forecast", fcstBuilder.Build())
		return nil
	})

	addAndHandleXRayRecordingError(ctx, err)
	return
}

func HandleLaunchRequest(ctx context.Context, request alexa.Request) (response alexa.Response) {
	err := xray.Capture(ctx, "LaunchRequestIntent", func(ctx1 context.Context) error {
		response = alexa.NewLaunchRequestGetPermissionsResponse()
		return nil
	})

	addAndHandleXRayRecordingError(ctx, err)
	return
}

func HandleHelpIntent(ctx context.Context, request alexa.Request) (response alexa.Response) {
	err := xray.Capture(ctx, "HelpIntent", func(ctx1 context.Context) error {
		var builder alexa.SSMLBuilder
		builder.Say("Here are some of the things you can ask:")
		builder.Pause("1000")
		builder.Say("Give me today's fishing forecast.")
		builder.Pause("1000")
		builder.Say("When is the best time to go fishing?")
		builder.Pause("1000")
		builder.Say("Get my fishing forecast please.")
		response = alexa.NewSSMLResponse("BugCaster Help", builder.Build())
		return nil
	})

	addAndHandleXRayRecordingError(ctx, err)
	return
}

func HandleAboutIntent(ctx context.Context, request alexa.Request) (response alexa.Response) {
	err := xray.Capture(ctx, "AboutIntent", func(ctx1 context.Context) error {
		var builder alexa.SSMLBuilder
		builder.Say("Welcome to Bug Caster!")
		builder.Pause("1000")
		builder.Say("Bug Caster uses solunar theory and applied analytics to determine how probable fish activity translates to quality of fishing by the hour.")
		builder.Pause("1000")
		builder.Say("Currently, ")
		builder.Pause("150")
		builder.Say("You can have Alexa ask Bug Caster for your fishing forecast, ")
		builder.Pause("150")
		builder.Say("or how the fishing is, ")
		builder.Pause("150")
		builder.Say("and get the best time to go fishing over the next couple of hours with a summarized rating and projected wind speed. ")
		builder.Pause("1000")
		builder.Say("New features will be coming soon, ")
		builder.Pause("150")
		builder.Say("including the ability to ask for a forecast for a specific time and location, ")
		builder.Pause("150")
		builder.Say("the best time during a specified range or normal daylight hours, ")
		builder.Pause("150")
		builder.Say("and potentially premium content such as a weekly forecast summary with graphic display. ")
		builder.Pause("1000")
		builder.Say("We hope Bug Caster improves your fishing experiences and appreciate any feedback through reviews on the skill page in the Alexa Skill Store! ")

		response = alexa.NewSSMLResponse("About BugCaster", builder.Build())
		return nil
	})

	addAndHandleXRayRecordingError(ctx, err)
	return
}

func Handler(ctx context.Context, request alexa.Request) (response alexa.Response, error error) {
	log.Print("Begin Handler")
	defer func() {
		// This shouldn't happen, but can if context gets cancelled without being properly handled.
		if r := recover(); r != nil {
			log.Print("Benzos delivered in Root Handler.")
			log.Print(r)
			response = alexa.NewDefaultErrorResponse()
		}
	}()

	return IntentDispatcher(ctx, request), nil
}

func main() {
	log.Print("Begin Main")
	lambda.Start(ContextConfigWrapper(Handler))
}
