package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type GeocodeService struct {
	URL           string
	UsrIp         string
	CountryRegion string
	PostalCode    string
	Key           string
	Client        http.Client
}

type GeocodeResponse struct {
	AuthenticationResultCode string           `json:"authenticationResultCode"`
	GeoResourceSets          []GeoResourceSet `json:"resourceSets"`
}

type GeoResourceSet struct {
	EstimatedTotal int           `json:"estimatedTotal"`
	GeoResources   []GeoResource `json:"resources"`
}

type GeoResource struct {
	Address struct {
		AdminDistrict    string
		AdminDistrict2   string
		CountryRegion    string
		FormattedAddress string
		Locality         string
		PostalCode       string
	} `json:"address"`
	Confidence        string   `json:"confidence"`
	EntityType        string   `json:"entityType"`
	GeoPoint          GeoPoint `json:"point"`
	StatusCode        int
	StatusDescription string
}

type GeoPoint struct {
	Coordinates []float64
}

func (s *GeocodeService) GetGeoPoint() (GeoPoint, error) {
	resp, err := s.GetAddressGeocodePoint()
	if err != nil {
		log.Printf("Failed to get Geocode Point: %s", err.Error())
		return GeoPoint{}, err
	}
	return resp.GeoResourceSets[0].GeoResources[0].GeoPoint, err
}

func (s *GeocodeService) GetAddressGeocodePoint() (*GeocodeResponse, error) {
	reqUrl := fmt.Sprintf("%s?countryRegion=%s&postalCode=%s&userIp=%s&key=%s", s.URL, s.CountryRegion, s.PostalCode, s.UsrIp, s.Key)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		log.Print("Error Creating Geo Request")
		return nil, err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Print("Error processing Geo Response")
		return nil, err
	}
	defer resp.Body.Close()

	geocodeResponse := GeocodeResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&geocodeResponse); err != nil {
		return nil, err
	}
	return &geocodeResponse, nil
}
