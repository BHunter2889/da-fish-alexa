package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"log"
	"net/http"
	"context"
	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/tomb.v2"
)

type DeviceService struct {
	URL      string
	Id       string
	Token    string
	Endpoint string
	t 		 tomb.Tomb
}

func (s *DeviceService) GetDeviceLocation(ctx context.Context) (*alexa.DeviceLocationResponse, error) {
	endp := fmt.Sprintf(s.Endpoint, s.Id)
	reqUrl := fmt.Sprintf("%s%s", s.URL, endp)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)

	if err != nil {
		log.Print("Error creating new device location request")
		log.Print(err)
		log.Print(req)
		return nil, err
	}
	bearer := "Bearer " + s.Token
	req.Header.Add("Authorization", bearer)

	resp, err := ctxhttp.Do(ctx, xray.Client(nil), req)
	if err != nil || resp.StatusCode == 403 {
		if resp.StatusCode == 403 {
			err = errors.New(resp.Status)
		}
		log.Print("Error processing device location response")
		log.Print(err)
		log.Print(resp)
		return nil, err
	}
	defer resp.Body.Close()

	deviceLocationResponse := alexa.DeviceLocationResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&deviceLocationResponse); err != nil {
		return nil, err
	}
	return &deviceLocationResponse, nil
}
