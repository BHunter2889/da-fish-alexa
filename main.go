package main

import (
	"context"
	"fmt"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"github.com/BHunter2889/da-fish-alexa/services"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
	"strings"
	"gopkg.in/tomb.v2"
)

var (
	cfg           *BugCasterConfig
	defaultUserIp = "127.0.0.1"

	deviceLocService  services.DeviceService
	GeocodeService    services.GeocodeService
	ForecasterService services.ForecasterService

	tombLoc *DeviceLocator
)

func killAllTombs() (err error) {
	if tombFR != nil {
		err = tombFR.Stop()
	}
	if tombGK != nil {
		err = tombGK.Stop()
	}
	if tombLoc != nil {
		err = tombLoc.Stop()
	}
	return
}

type DeviceLocator struct {
	ctx              context.Context
	DeviceLocService *services.DeviceService
	Ch               chan *alexa.DeviceLocationResponse
	ErCh             chan error
	t                tomb.Tomb
}

func NewDeviceLocator(ctx context.Context, ds *services.DeviceService) *DeviceLocator {
	dl := &DeviceLocator{
		ctx:              ctx,
		DeviceLocService: ds,
		Ch:               make(chan *alexa.DeviceLocationResponse),
		ErCh:             make(chan error),
	}
	dl.t.Go(dl.getDeviceLocation)
	return dl
}

func (dl *DeviceLocator) Stop() error {
	dl.t.Kill(nil)
	return dl.t.Wait()
}

func (dl *DeviceLocator) getDeviceLocation() error {
	resp, err := deviceLocService.GetDeviceLocation(dl.ctx)
	if err != nil {
		log.Print(resp)
		log.Print(err)
		if strings.Contains(err.Error(), "403") {
			dl.ErCh <- err
			close(dl.Ch)
			close(dl.ErCh)
			return nil
		}
		close(dl.Ch)
		close(dl.ErCh)
		return err
	}

	select {
	case dl.Ch <- resp:
		close(dl.ErCh)
		return nil
	case <-dl.t.Dying():
		log.Print("DeviceLocator Dying...")
		close(dl.Ch)
		close(dl.ErCh)
		return nil
	}
}

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
		// Kill Running processes, print the error, move on to returning the default response.
		err := killAllTombs()
		log.Print(err)
		response = HandleAboutIntent(ctx, request)
	}
	return response
}

func HandleTodaysFishRatingIntent(ctx context.Context, request alexa.Request) (response alexa.Response) {
	log.Print("Todays Fish Rating Intent")

	// Handle Any Panics with a Default Error Response
	defer func() {
		if r := recover(); r != nil {
			log.Print("Benzos delivered in TodaysFishRatingIntent handler.")
			log.Print(r)
			response = alexa.NewDefaultErrorResponse()
		}
	}()

	deviceId := request.Context.System.Device.DeviceID
	apiAccessToken := request.Context.System.APIAccessToken
	apiEndpoint := request.Context.System.APIEndpoint

	// Get Location registered to user device
	deviceLocService = services.DeviceService{
		URL:      apiEndpoint,
		Id:       deviceId,
		Token:    apiAccessToken,
		Endpoint: cfg.AlexaLocEndpoint,
	}

	tombLoc := NewDeviceLocator(ctx, &deviceLocService)
	var err error
	var resp *alexa.DeviceLocationResponse

	// Listen to select either a Successful Response, or Error, whichever comes first.
	select {
	case err = <-tombLoc.ErCh:
		if err != nil {

			// Expected case when Location Permissions have not yet been enabled.
			if strings.Contains(err.Error(), "403") {
				log.Print("REQUESTING LOCATION PERMISSIONS")

				// Kill any processes still running, Log the Error, return new Permissions Request Response
				err := killAllTombs()
				log.Print(err)
				return alexa.NewPermissionsRequestResponse()
			} else { // Trouble Communicating with Alexa Devices Api
				log.Print("Unexpected Device Location Retrieval Error")

				// Kill Processes, return fallback response.
				err := killAllTombs()
				log.Print(err)
				panic("Unexpected Device Location Retrieval Error")
			}
		}
	case resp = <-tombLoc.Ch:
		log.Print(resp)
	}

	// Wait for KMS decryption service calls to finish if they haven't before proceeding.
	KMSDecrytiponWaiter()

	// Get Geocode coordinates from retrieved location
	GeocodeService = services.GeocodeService{
		URL:           cfg.GeoUrl,
		UsrIp:         defaultUserIp,
		CountryRegion: resp.CountryCode,
		PostalCode:    resp.PostalCode,
		Key:           cfg.GeoKey,
		Client:        http.Client{},
	}
	log.Print(cfg.GeoKey)
	log.Print(cfg.FishRatingUrl)
	log.Print("Geocoding Location... ")

	geoPoint, err := GeocodeService.GetGeoPoint(ctx)
	if err != nil {
		log.Print("trouble getting information on the area ")
		log.Print(err)
		panic("trouble getting information on the area ")
	}
	log.Printf("Geo: {lat: %f, lon: %f}", geoPoint.Coordinates[0], geoPoint.Coordinates[1])

	//Get Fishing Forecast using coordinates
	ForecasterService = services.ForecasterService{
		URL:    cfg.FishRatingUrl,
		Lat:    geoPoint.Coordinates[0],
		Lon:    geoPoint.Coordinates[1],
		Client: http.Client{},
	}
	log.Print("Getting Forecast... ")
	fr, err := ForecasterService.GetCurrentFishingRating(ctx)
	if err != nil {
		log.Print(err)
		log.Print("Trouble Getting Fishing Forecast ")
		panic("Trouble Getting Fishing Forecast ")
	}
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

	return alexa.NewSSMLResponse("Today's Fishing Forecast", fcstBuilder.Build())
}

func HandleLaunchRequest(request alexa.Request) alexa.Response {
	return alexa.NewLaunchRequestGetPermissionsResponse()
}

func HandleHelpIntent(ctx context.Context, request alexa.Request) alexa.Response {
	var builder alexa.SSMLBuilder
	builder.Say("Here are some of the things you can ask:")
	builder.Pause("1000")
	builder.Say("Give me today's fishing forecast.")
	builder.Pause("1000")
	builder.Say("When is the best time to go fishing?")
	builder.Pause("1000")
	builder.Say("Get my fishing forecast please.")
	return alexa.NewSSMLResponse("BugCaster Help", builder.Build())
}

func HandleAboutIntent(ctx context.Context, request alexa.Request) alexa.Response {
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

	return alexa.NewSSMLResponse("About BugCaster", builder.Build())
}

func Handler(ctx context.Context, request alexa.Request) (response alexa.Response, error error) {
	log.Print("Begin Handler")
	defer func() {
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
