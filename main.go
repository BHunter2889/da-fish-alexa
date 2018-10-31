package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"github.com/BHunter2889/da-fish-alexa/services"
	"net/http"
	"log"
	"fmt"
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

func HandleTodaysFishRatingIntent(request alexa.Request) alexa.Response {
	defer func() alexa.Response {
		var resp alexa.Response
		if r := recover(); r != nil {
			resp = alexa.NewSimpleResponse("Today's Fishing Forecast", "You caught me! Like a young fish, I'm still learning. "+
				"Please be patient with me, I'll have forecasts for you soon!")
		}

		return resp;
	}()

	deviceId := request.Context.System.Device.DeviceID
	apiAccessToken := request.Context.System.APIAccessToken
	apiEndpoint := request.Context.System.APIEndpoint

	// Get Location registered to user device
	DeviceLocService = services.DeviceService{
		URL:    apiEndpoint,
		Id:     deviceId,
		Token:  apiAccessToken,
		Client: http.Client{},
	}

	resp, err := DeviceLocService.GetDeviceLocation()
	if err != nil {
		panic("trouble getting your zip code area, please ensure you have granted permission for this so that " +
			"Da Fish can determine the fishing forecast for your area.")
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
		r uint
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
		fcstBuilder.Pause("250")
		fcstBuilder.Say("well, ")
		fcstBuilder.Pause("250")
		fcstBuilder.Say(
			fmt.Sprintf("probably some other time with a top rating well below average and a wind speed of %f.", w))
	} else if r >= 3 && r <= 4 {
		fcstBuilder.Say(fmt.Sprintf("It looks like a decent or possibly better time to go fishing %s with a forecast rating "+
			"just on the plus side.", t))
		fcstBuilder.Pause("500")
		fcstBuilder.Say(fmt.Sprintf("The wind speed is listed at %f.", w))
	} else {
		fcstBuilder.Say("The fish appear to be biting!")
		fcstBuilder.Pause("750")
		fcstBuilder.Say(fmt.Sprintf("Over the next couple hours the fishing looks great "+
			"with the best time to go being %s with a forecast rating well over the norm. ", t))
		fcstBuilder.Pause("500")
		fcstBuilder.Say(fmt.Sprintf("The wind speed is listed at %f.", w))
	}

	return alexa.NewSSMLResponse("Today's Fishing Forecast", fcstBuilder.Build())
}

func HandleHelpIntent(request alexa.Request) alexa.Response {
	var builder alexa.SSMLBuilder
	builder.Say("Here are some of the things you can ask:")
	builder.Pause("1000")
	builder.Say("Give me todays fishing forecast.")
	builder.Pause("1000")
	builder.Say("When is the best time to go fishing?")
	return alexa.NewSSMLResponse("Da Fish Forecaster Help", builder.Build())
}

func HandleAboutIntent(request alexa.Request) alexa.Response {
	return alexa.NewSimpleResponse("About", "Da Fish was created by HuntX in Saint Louis, Missouri so that he couldn't talk himself out of going fishing by using the excuse that conditions may not be optimal and figuring it out takes too much time to look up.")
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

////TODO Remove
//func HandleFrontpageDealIntent(request alexa.Request) alexa.Response {
//	feedResponse, _ := RequestFeed("frontpage")
//	var builder alexa.SSMLBuilder
//	builder.Say("Here are the current frontpage deals:")
//	builder.Pause("1000")
//	for _, item := range feedResponse.Channel.Item {
//		builder.Say(item.Title)
//		builder.Pause("1000")
//	}
//	return alexa.NewSSMLResponse("Frontpage Deals", builder.Build())
//}
//
////TODO Remove
//func HandlePopularDealIntent(request alexa.Request) alexa.Response {
//	return alexa.NewSimpleResponse("Popular Deals", "Popular deal data here")
//}
//// TODO - Delete this or rework for JSON
//type FeedResponse struct {
//	Channel struct {
//		Item []struct {
//			Title string `xml:"title"`
//			Link  string `xml:"link"`
//		} `xml:"item"`
//	} `xml:"channel"`
//}
//
//func RequestFeed(mode string) (FeedResponse, error) {
//	endpoint, _ := url.Parse("https://slickdeals.net/newsearch.php")
//	queryParams := endpoint.Query()
//	queryParams.Set("mode", mode)
//	queryParams.Set("searcharea", "deals")
//	queryParams.Set("searchin", "first")
//	queryParams.Set("rss", "1")
//	endpoint.RawQuery = queryParams.Encode()
//	response, err := http.Get(endpoint.String())
//	if err != nil {
//		return FeedResponse{}, err
//	} else {
//		data, _ := ioutil.ReadAll(response.Body)
//		var feedResponse FeedResponse
//		xml.Unmarshal(data, &feedResponse)
//		return feedResponse, nil
//	}
//}
