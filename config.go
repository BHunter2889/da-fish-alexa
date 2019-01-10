package main

import (
	"context"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-xray-sdk-go/xray"
	"gopkg.in/tomb.v2"
	"log"
	"os"
	"sync"
)

type BugCasterConfig struct {
	AlexaApiUrl      string
	AlexaLocEndpoint string
	GeoKey           string
	GeoUrl           string
	FishRatingUrl    string
	ImageUrls        struct {
		BgImageMedPos1 string
		BgImageMedPos2 string
		BgImageMedNeg1 string
		BugCasterLogo  string
	}
	APLDirectiveTemplate alexa.Directive
	t                    tomb.Tomb
}

// Defining as constants rather than reading from config file - maybe text w/ X-ray to see how much longer reading from
// a config file or otherwise might take.
// AlexaApiBaseUrl  = "https://api.amazonalexa.com"  --- US Endpoint. Will populate this dynamically from the incoming Request payload.
// %s - reserved for DeviceId
const AlexaLocEndpoint = "/v1/devices/%s/settings/address/countryAndPostalCode"

// Filename for the APL Document to read
const AplDocFilename = "bugcaster-apl-doc-2.json"

var (
	kMS        *kms.KMS
	sess       = session.Must(session.NewSession())
	wg         sync.WaitGroup
	chanFR     <-chan string
	chanGK     <-chan string
	tombFR     *KMSDecryptTomb
	tombGK     *KMSDecryptTomb
	t          tomb.Tomb
	supportAPL = false
)

type AlexaRequestHandler func(context.Context, alexa.Request) (alexa.Response, error)

// Wrap The Handler so that we can use context to do some config BEFORE proceeding with handler.
func ContextConfigWrapper(h AlexaRequestHandler) AlexaRequestHandler {
	return func(ctx context.Context, request alexa.Request) (response alexa.Response, err error) {
		log.Print(request)

		if &request.Context.Display != nil {
			supportAPL = true
		}

		// Put up a Border Wall (which they can very easily get around)
		if request.Body.Locale != "en-US" && request.Body.Locale != "en-CA" {
			return alexa.NewUnsupportedLocationResponse(), nil
		}

		// If this is a Launch Request, we don't need Config at all, so kick it back out before it causes problems
		if request.Body.Type == "LaunchRequest" {
			return HandleLaunchRequest(ctx, request), nil
		}

		// Benzos PRN - Take once at bedtime as needed. (Defer a panic resolution which returns a default error voice response to the user.)
		defer func() {
			if r := recover(); r != nil {
				log.Print("CONTEXT WRAPPER PANIC")
				log.Print(err)
				log.Print(r)
				response = alexa.NewDefaultErrorResponse()
			}
		}()

		// Logging Context for Demo Purposes
		log.Print(ctx)
		addAndHandleXRayRecordingError(ctx, xray.AddMetadata(ctx, "lambda-context", ctx))

		if err := NewBugCasterConfig(ctx); err != nil {
			log.Print(err)
			panic(err)
		}

		response, err = h(ctx, request)
		if err != nil {
			log.Print(err)
			panic(err.Error())
		}
		log.Print(response)
		return response, nil
	}
}

func NewKMS() *kms.KMS {
	log.Print("Init kMS Config")
	c := kms.New(sess)
	xray.AWS(c.Client)
	return c
}

func NewBugCasterConfig(ctx context.Context) (err1 error) {
	// Record Config Performance Impact and Profile Errors.
	err := xray.Capture(ctx, "config.New", func(ctx1 context.Context) error {
		log.Print("NewBugCasterConfig")
		kMS = NewKMS()
		cfg = new(BugCasterConfig)
		err1 = cfg.LoadConfig(ctx)
		return nil
	})

	addAndHandleXRayRecordingError(ctx, err)
	return
}

func KMSDecryptionWaiter() {
	//log.Print("Waiting on kMS Decryption...")
	cfg.FishRatingUrl = <-tombFR.Ch
	//log.Printf("FRU: %s", cfg.FishRatingUrl)
	cfg.GeoKey = <-tombGK.Ch
	//log.Printf("GK: %s", cfg.GeoKey)
	wg.Wait()
	//log.Print("Done Waiting On kMS Decryption.")
}

func init() {
	//log.Print("Init Xray in Config")
	err := xray.Configure(xray.Config{
		LogLevel: "info",
	})
	log.Print(err)
}

func (cfg *BugCasterConfig) LoadConfig(ctx context.Context) (err error) {
	//log.Print("Begin LoadConfig")
	wg.Add(2)
	cfg.AlexaLocEndpoint = AlexaLocEndpoint

	// Start a new KMSDecryption X-Ray Subsegment to evaluate performance
	addAndHandleXRayRecordingError(ctx, xray.Capture(ctx, "KMSDecryption", func(ctx1 context.Context) error {
		tombFR = NewKMSDecryptTomb(ctx, "FISH_RATING_SERVICE_URL")
		tombGK = NewKMSDecryptTomb(ctx, "GEO_KEY")
		return nil
	}))
	cfg.GeoUrl = os.Getenv("GEO_SERVICE_URL")

	if supportAPL {
		cfg.ImageUrls.BgImageMedPos1 = os.Getenv("BG_IMAGE_MED_POS_1")
		cfg.ImageUrls.BgImageMedPos2 = os.Getenv("BG_IMAGE_MED_POS_2")
		cfg.ImageUrls.BgImageMedNeg1 = os.Getenv("BG_IMAGE_MED_NEG_1")
		cfg.ImageUrls.BugCasterLogo = os.Getenv("BUGCASTER_LOGO")

		// Consider Adding X-Ray Support to this...?
		go func() {
			rd := alexa.Directive{}
			if err := alexa.ExtractNewRenderDocDirectiveFromString("testing", aplJson, &rd); err != nil {
				log.Print(err)
			}
			cfg.APLDirectiveTemplate = rd
		}()
	}
	return
}

func addAndHandleXRayRecordingError(ctx context.Context, err error) {
	if err != nil {
		log.Print(err)
		if err1 := xray.AddError(ctx, err); err1 != nil {
			log.Print(err1)
		}
	}
}
