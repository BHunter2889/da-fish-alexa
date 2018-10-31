package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/BHunter2889/da-fish/alexa"
	"github.com/BHunter2889/da-fish/services"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/xml"
	"log"
)
// TODO - add context.Context for xray tracing

var (
	cfg = DaFishConfig{}
	defaultUserIp = "127.0.0.1"

	DeviceLocService services.DeviceService
	GeocodeService	services.GeocodeService
)

func IntentDispatcher(request alexa.Request) alexa.Response {
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "TodaysFishRatingIntent":
		response = HandleTodaysFishRatingIntent(request)
	case "FrontpageDealIntent":
		response = HandleFrontpageDealIntent(request)
	case "PopularDealIntent":
		response = HandlePopularDealIntent(request)
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

	resp, _ := DeviceLocService.GetDeviceLocation()
	log.Print(resp)

	// Get Geocode coordinates from retrieved location
	GeocodeService = services.GeocodeService{
		URL: cfg.GeoUrl,
		UsrIp: defaultUserIp,
		CountryRegion: resp.CountryCode,
		PostalCode: resp.PostalCode,
		Key: cfg.GeoKey,
	}

	geoPoint, _ := GeocodeService.GetGeoPoint()
	log.Printf("Geo: {lat: %f, lon: %f}", geoPoint.Coordinates[0], geoPoint.Coordinates[1])

	//Get Fishing Forecast using coordinates
	//TODO - Start Here,  use GeoPoint to get  FishRating

	return alexa.NewSimpleResponse("Today's Fishing Forecast", "You caught me! Like a young fish, I'm still learning. "+
		"Please be patient with me, I'll have forecasts for you soon!")
}

//TODO Remove
func HandleFrontpageDealIntent(request alexa.Request) alexa.Response {
	feedResponse, _ := RequestFeed("frontpage")
	var builder alexa.SSMLBuilder
	builder.Say("Here are the current frontpage deals:")
	builder.Pause("1000")
	for _, item := range feedResponse.Channel.Item {
		builder.Say(item.Title)
		builder.Pause("1000")
	}
	return alexa.NewSSMLResponse("Frontpage Deals", builder.Build())
}

//TODO Remove
func HandlePopularDealIntent(request alexa.Request) alexa.Response {
	return alexa.NewSimpleResponse("Popular Deals", "Popular deal data here")
}

func HandleHelpIntent(request alexa.Request) alexa.Response {
	// TODO
	var builder alexa.SSMLBuilder
	builder.Say("Here are some of the things you can ask:")
	builder.Pause("1000")
	builder.Say("Give me the frontpage deals.")
	builder.Pause("1000")
	builder.Say("Give me the popular deals.")
	return alexa.NewSSMLResponse("Slick Dealer Help", builder.Build())
}

func HandleAboutIntent(request alexa.Request) alexa.Response {
	return alexa.NewSimpleResponse("About", "Da Fish was created by HuntX in Saint Louis, Missouri so that he couldn't talk himself out of going fishing by using the excuse that conditions may not be optimal and figuring it out takes too much time to look up.")
}

// TODO - Delete this or rework for JSON
type FeedResponse struct {
	Channel struct {
		Item []struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
		} `xml:"item"`
	} `xml:"channel"`
}

func RequestFeed(mode string) (FeedResponse, error) {
	endpoint, _ := url.Parse("https://slickdeals.net/newsearch.php")
	queryParams := endpoint.Query()
	queryParams.Set("mode", mode)
	queryParams.Set("searcharea", "deals")
	queryParams.Set("searchin", "first")
	queryParams.Set("rss", "1")
	endpoint.RawQuery = queryParams.Encode()
	response, err := http.Get(endpoint.String())
	if err != nil {
		return FeedResponse{}, err
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var feedResponse FeedResponse
		xml.Unmarshal(data, &feedResponse)
		return feedResponse, nil
	}
}

func Handler(request alexa.Request) (alexa.Response, error) {
	return IntentDispatcher(request), nil
}

// Load Properties before proceeding
func init() {
	cfg.LoadConfig()
}
func main() {
	lambda.Start(Handler)
}
