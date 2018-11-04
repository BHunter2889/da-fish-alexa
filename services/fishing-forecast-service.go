package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ForecasterService struct {
	URL    string
	Lat    float64
	Lon    float64
	Client http.Client
}
type ForecasterResponse struct {
	Currently struct {
		ApparentTemperature  float64 `json:"apparentTemperature"`
		CloudCover           float64 `json:"cloudCover"`
		DewPoint             float64 `json:"dewPoint"`
		Humidity             float64 `json:"humidity"`
		Icon                 string  `json:"icon"`
		NearestStormBearing  int     `json:"nearestStormBearing"`
		NearestStormDistance int     `json:"nearestStormDistance"`
		Ozone                float64 `json:"ozone"`
		PrecipIntensity      float64 `json:"precipIntensity"`
		PrecipProbability    float64 `json:"precipProbability"`
		//pressure: 1009.26
		Summary     string  `json:"summary"`
		Temperature float64 `json:"temperature"`
		Time        int64
		//uvIndex: 0
		//visibility: 10
		WindBearing int     `json:"windBearing"`
		WindGust    float64 `json:"windGust"`
		WindSpeed   float64 `json:"windSpeed"`
	} `json:"currently"`
	Hourly []Hour `json:"hourly"`
}

type Hour struct {
	Icon              string  `json:"icon"`
	PrecipProbability float64 `json:"precipProbability"`
	Rating            int    `json:"rating"`
	Temperature       float64 `json:"temperature"`
	Time              int64   `json:"time"`
	WindSpeed         float64 `json:"windSpeed"`
}

func (s *ForecasterService) GetCurrentFishingRating() ([]Hour, error) {
	resp, err := s.GetFullFishingForecast()
	if err != nil {
		log.Printf("Failed to get Fishing Forecast: %s", err.Error())
		return []Hour{}, err
	}
	//For now Let's just return now and 2 hours from now.
	log.Print(resp)
	log.Print(resp.Hourly)
	response := []Hour{resp.Hourly[0], resp.Hourly[2]}
	return response, nil
}

func (s *ForecasterService) GetFullFishingForecast() (*ForecasterResponse, error) {
	reqUrl := fmt.Sprintf("%s?lat=%f&lon=%f", s.URL, s.Lat, s.Lon)
	log.Printf("Fishing Forecast Request: %s", reqUrl)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		log.Print("Error creating Fishing Forecast Request")
		log.Print(err)
		return nil, err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Print("Error processing Fishing Forecast Response")
		log.Print(err)
		return nil, err
	}
	defer resp.Body.Close()

	forecasterResponse := ForecasterResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&forecasterResponse); err != nil {
		return nil, err
	}
	return &forecasterResponse, nil
}
