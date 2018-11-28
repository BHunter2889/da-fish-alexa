package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/net/context/ctxhttp"
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

func (s *GeocodeService) GetGeoPoint(ctx context.Context) (GeoPoint, error) {
	resp, err := s.GetAddressGeocodePoint(ctx)
	if err != nil {
		log.Printf("Failed to get Geocode Point: %s", err.Error())
		return GeoPoint{}, err
	}
	return resp.GeoResourceSets[0].GeoResources[0].GeoPoint, nil
}

func (s *GeocodeService) GetAddressGeocodePoint(ctx context.Context) (GeocodeResponse, error) {
	reqUrl := fmt.Sprintf("%s?countryRegion=%s&postalCode=%s&userIp=%s&key=%s", s.URL, s.CountryRegion, s.PostalCode, s.UsrIp, s.Key)
	resp, err := ctxhttp.Get(ctx, xray.Client(nil), reqUrl)

	if err != nil {
		log.Print("Error processing Geo Response")
		return GeocodeResponse{}, err
	}
	defer resp.Body.Close()

	geocodeResponse := GeocodeResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&geocodeResponse); err != nil {
		return GeocodeResponse{}, err
	}
	return geocodeResponse, nil
}
