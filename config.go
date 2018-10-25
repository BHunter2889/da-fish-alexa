package da_fish

import (
	"os"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/base64"
	"github.com/aws/aws-sdk-go/service/kms"
)

type Config struct {
	AlexaApiUrl           string
	AlexaLocationEndpoint string
	GeoKey                string
	GeoUrl                string
	FishRatingUrl         string
}

var geoKeyEncrypt string = os.Getenv("GEO_KEY")
var fishRatingUrlEncrypt string = os.Getenv("FISH_RATING_SERVICE_URL")
var geoUrl string = os.Getenv("GEO_SERVICE_URL")

var geoKeyDecrypt string
var fishRatingUrlDecrypt string

func init() {
	kmsClient := kms.New(session.New())
	//TODO
	decodedBytes, err := base64.StdEncoding.DecodeString(geoKeyDecrypt)
	if err != nil {
		panic(err)
	}
	input := &kms.DecryptInput{
		CiphertextBlob: decodedBytes,
	}
	response, err := kmsClient.Decrypt(input)
	if err != nil {
		panic(err)
	}
	// Plaintext is a byte array, so convert to string
	geoKeyDecrypt = string(response.Plaintext[:])
}