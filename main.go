package main

import (
	"fmt"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"github.com/BHunter2889/da-fish-alexa/services"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
	"strings"
)

// TODO - add context.Context for xray tracing

var (
	cfg           *DaFishConfig
	defaultUserIp = "127.0.0.1"

	DeviceLocService  services.DeviceService
	GeocodeService    services.GeocodeService
	ForecasterService services.ForecasterService
)

func IntentDispatcher(request alexa.Request) alexa.Response {
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "TodaysFishRatingIntent":
		response = HandleTodaysFishRatingIntent(request)
		//case "FrontpageDealIntent":
		//	response = HandleFrontpageDealIntent(request)
		//case "PopularDealIntent":
		//	response = HandlePopularDealIntent(request)
	case alexa.HelpIntent:
		response = HandleHelpIntent(request)
	case "AboutIntent":
		response = HandleAboutIntent(request)
	default:
		response = HandleAboutIntent(request)
	}
	return response
}

func HandleTodaysFishRatingIntent(request alexa.Request) (response alexa.Response) {
	defer func() {
		if r := recover(); r != nil {
			response = alexa.NewDefaultErrorResponse()
		}
	}()

	deviceId := request.Context.System.Device.DeviceID
	apiAccessToken := request.Context.System.APIAccessToken
	apiEndpoint := request.Context.System.APIEndpoint

	log.Printf("Device ID: %s, ApiAccess: %s, Endpoint: %s", deviceId, apiAccessToken, apiEndpoint)

	// Get Location registered to user device
	DeviceLocService = services.DeviceService{
		URL:      apiEndpoint,
		Id:       deviceId,
		Token:    apiAccessToken,
		Endpoint: cfg.AlexaLocEndpoint,
		Client:   http.Client{},
	}

	resp, err := DeviceLocService.GetDeviceLocation()
	if err != nil {
		log.Print(resp)
		log.Print(err)
		// TODO - Consider adding custom prompt if possible
		//var builder = alexa.SSMLBuilder{}
		if strings.Contains(err.Error(), "403") {
			return alexa.NewPermissionsRequestResponse()
		} else {
			panic("Unexpected Device Location Retrieval Error")
		}
	}
	log.Print(resp)
	// Get Geocode coordinates from retrieved location
	GeocodeService = services.GeocodeService{
		URL:           cfg.GeoUrl,
		UsrIp:         defaultUserIp,
		CountryRegion: resp.CountryCode,
		PostalCode:    resp.PostalCode,
		Key:           cfg.GeoKey,
		Client:        http.Client{},
	}

	log.Printf("Geo URL: %s", GeocodeService.URL)
	geoPoint, err := GeocodeService.GetGeoPoint()
	if err != nil {
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

	fr, err := ForecasterService.GetCurrentFishingRating()
	if err != nil {
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

	var fcstBuilder alexa.SSMLBuilder
	if r < 3 {
		fcstBuilder.Say("It looks like the best time to go fishing over the next couple of hours is, ")
		fcstBuilder.Pause("150")
		fcstBuilder.Say("well, ")
		fcstBuilder.Pause("150")
		fcstBuilder.Say("probably some other time.")
		fcstBuilder.Pause("500")
		fcstBuilder.Say(fmt.Sprintf(" The top rating is well below average and the wind speed is %.1f miles per hour.", w))
	} else if r >= 3 && r <= 4 {
		fcstBuilder.Say(fmt.Sprintf("It looks like a decent or possibly better time to go fishing %s.", t))
		fcstBuilder.Say(" The forecast rating just on the plus side,")
		fcstBuilder.Pause("250")
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

func HandleHelpIntent(request alexa.Request) alexa.Response {
	var builder alexa.SSMLBuilder
	builder.Say("Here are some of the things you can ask:")
	builder.Pause("1000")
	builder.Say("Give me today's fishing forecast.")
	builder.Pause("1000")
	builder.Say("When is the best time to go fishing?")
	return alexa.NewSSMLResponse("BugCaster Help", builder.Build())
}

func HandleAboutIntent(request alexa.Request) alexa.Response {
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
	builder.Say("We hope Bug Caster improves your fishing experiences and appreciate any feedback! ")

	return alexa.NewSSMLResponse("About BugCaster", builder.Build())
}

func Handler(request alexa.Request) (alexa.Response, error) {
	return IntentDispatcher(request), nil
}

// Load Properties before proceeding
func init() {
	cfg = new(DaFishConfig)
	cfg.LoadConfig()
}
func main() {
	lambda.Start(Handler)
}
