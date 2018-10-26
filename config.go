package da_fish

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/base64"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-xray-sdk-go/xray"
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
const (
	AlexaLocEndpoint = "/v1/devices/*deviceId*/settings/address/countryAndPostalCode"
	AlexaApiBaseUrl  = "https://api.amazonalexa.com"
)

var (
	KMS    = NewKMS()
	sess   = session.Must(session.NewSession())
	config = new(DaFishConfig)

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
func NewKMS() *kms.KMS {
	c := kms.New(sess)
	xray.AWS(c.Client)
	return c
}

func (cfg *DaFishConfig) LoadConfig() {
	cfg.AlexaApiUrl = AlexaApiBaseUrl
	cfg.AlexaLocEndpoint = AlexaLocEndpoint
	cfg.GeoKey = geoKeyDecrypt
	cfg.GeoUrl = geoUrl
	cfg.FishRatingUrl = fishRatingUrlDecrypt
}
