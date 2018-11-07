package main

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	//"github.com/aws/aws-xray-sdk-go/xray"
	"os"
)

type DaFishConfig struct {
	AlexaApiUrl      string
	AlexaLocEndpoint string
	GeoKey           string
	GeoUrl           string
	FishRatingUrl    string
}

// Defining as constants rather than reading from config file until resource monitoring is setup
// AlexaApiBaseUrl  = "https://api.amazonalexa.com"  --- US Endpoint. Will grab this from the incoming Request payload.
// %s - reserved for DeviceId
const AlexaLocEndpoint = "/v1/devices/%s/settings/address/countryAndPostalCode"

var (
	KMS  = NewKMS()

	sess = session.Must(session.NewSession())

	geoKeyDecrypt        = decrypt(os.Getenv("GEO_KEY"))
	fishRatingUrlDecrypt = decrypt(os.Getenv("FISH_RATING_SERVICE_URL"))
	geoUrl               = os.Getenv("GEO_SERVICE_URL")
)

func decrypt(s string) string {
	decodedBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	input := &kms.DecryptInput{
		CiphertextBlob: decodedBytes,
	}
	response, err := KMS.Decrypt(input)
	if err != nil {
		panic(err)
	}
	// Plaintext is a byte array, so convert to string
	return string(response.Plaintext[:])
}

// Wrap in Xray so we can detail any errors
// TODO - Fix Xray
func NewKMS() *kms.KMS {
	c := kms.New(sess)
	//xray.AWS(c.Client)
	return c
}

// TODO - finish Xray impl
//func init() {
//	xray.Configure(xray.Config{
//		LogLevel: "trace",
//	})
//}


func (cfg *DaFishConfig) LoadConfig() {
	cfg.AlexaLocEndpoint = AlexaLocEndpoint
	cfg.GeoKey = geoKeyDecrypt
	cfg.GeoUrl = geoUrl
	cfg.FishRatingUrl = fishRatingUrlDecrypt
}
