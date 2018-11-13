package main

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-xray-sdk-go/xray"
	"os"
	"log"
	"context"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"sync"
)

type BugCasterConfig struct {
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
	KMS *kms.KMS
	sess = session.Must(session.NewSession())
	wg  sync.WaitGroup
	chanFR <- chan string
	chanGK <- chan string
)

type AlexaRequestHandler func(context.Context, alexa.Request) (alexa.Response, error)

// Wrap The Handler so that we can use context to do some config BEFORE proceeding with handler.
func ContextConfigWrapper(h AlexaRequestHandler) AlexaRequestHandler {
	return func(ctx context.Context, request alexa.Request) (alexa.Response, error) {
		log.Print(ctx)
		<- NewBugCasterConfig(ctx)
		return h(ctx, request)
	}
}

// We want this in a channel
func decrypt(ctx context.Context, s string) (wait <- chan string) {
	log.Print("New Decrypt...")
	ch := make(chan string)
	go func() {
		decodedBytes, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			panic(err)
		}
		input := &kms.DecryptInput{
			CiphertextBlob: decodedBytes,
		}
		response, err := KMS.DecryptWithContext(ctx, input)
		if err != nil {
			panic(err)
		}
		// Plaintext is a byte array, so convert to string
		ch <- string(response.Plaintext[:])
		close(ch)
		wg.Done()
	}()
	return ch
}

// Wrap in Xray so we can detail any errors
// TODO - Fix Xray
func NewKMS() *kms.KMS {
	log.Print("Init KMS Config")
	c := kms.New(sess)
	xray.AWS(c.Client)
	return c
}

func NewBugCasterConfig(ctx context.Context) (wait <- chan struct{}) {
	log.Print("NewBugCasterConfig")
	ch := make(chan struct{})
	go func() {
		KMS = NewKMS()
		cfg = new(BugCasterConfig)
		cfg.LoadConfig(ctx)
		close(ch)
	}()
	return ch
}

func KMSDecrytiponWaiter() {
	cfg.FishRatingUrl = <- chanFR
	cfg.GeoKey = <- chanGK
	wg.Wait()
}

func init() {
	log.Print("Init Xray in Config")
	xray.Configure(xray.Config{
		LogLevel: "trace",
	})
}


func (cfg *BugCasterConfig) LoadConfig(ctx context.Context) {
	log.Print("Begin LoadConfig")
	wg.Add(2)
	cfg.AlexaLocEndpoint = AlexaLocEndpoint
	chanFR = decrypt(ctx, os.Getenv("FISH_RATING_SERVICE_URL"))
	chanGK = decrypt(ctx, os.Getenv("GEO_KEY"))
	cfg.GeoUrl = os.Getenv("GEO_SERVICE_URL")
}
