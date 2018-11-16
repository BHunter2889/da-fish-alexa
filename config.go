package main

import (
	"context"
	"encoding/base64"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-xray-sdk-go/xray"
	"log"
	"os"
	"sync"
	"gopkg.in/tomb.v2"
)

type BugCasterConfig struct {
	AlexaApiUrl      string
	AlexaLocEndpoint string
	GeoKey           string
	GeoUrl           string
	FishRatingUrl    string
	t                tomb.Tomb
}

type KMSDecryptTomb struct {
	ctx context.Context
	s   string
	Ch  chan string
	t   tomb.Tomb
}

// Defining as constants rather than reading from config file until resource monitoring is setup
// AlexaApiBaseUrl  = "https://api.amazonalexa.com"  --- US Endpoint. Will grab this from the incoming Request payload.
// %s - reserved for DeviceId
const AlexaLocEndpoint = "/v1/devices/%s/settings/address/countryAndPostalCode"

var (
	KMS    *kms.KMS
	sess   = session.Must(session.NewSession())
	wg     sync.WaitGroup
	chanFR <-chan string
	chanGK <-chan string
	tombFR *KMSDecryptTomb
	tombGK *KMSDecryptTomb
	t      tomb.Tomb
)

type AlexaRequestHandler func(context.Context, alexa.Request) (alexa.Response, error)

// Wrap The Handler so that we can use context to do some config BEFORE proceeding with handler.
func ContextConfigWrapper(h AlexaRequestHandler) AlexaRequestHandler {
	return func(ctx context.Context, request alexa.Request) (response alexa.Response, err error) {
		log.Print(request)

		// Put up a Border Wall (which they can very easily get around)
		if request.Body.Locale != "en-US" && request.Body.Locale != "en-CA" {
			return alexa.NewUnsupportedLocationResponse(), nil
		}

		// If this is a Launch Request, we don't need Config at all, so kick it back out before it causes problems
		if request.Body.Type == "LaunchRequest" {
			return HandleLaunchRequest(request), nil
		}

		defer func() {
			if r := recover(); r != nil {
				log.Print("CONTEXT WRAPPER PANIC")
				log.Print(err)
				log.Print(r)
				response = alexa.NewDefaultErrorResponse()
			}
		}()
		log.Print(ctx)
		NewBugCasterConfig(ctx)

		response, err = h(ctx, request)
		if err != nil {
			log.Print(err)
			panic(err.Error())
		}
		log.Print(response)
		return response, nil
	}
}

// We want this in a channel
// Logging For Demo Purposes
func (kdt *KMSDecryptTomb) decrypt() error {
	log.Print("New Decrypt...")
	//go func() {
	log.Print("Go Decrypt...")
	decodedBytes, err := base64.StdEncoding.DecodeString(kdt.s)
	if err != nil {
		log.Print(err)
		if err.Error() == request.CanceledErrorCode {
			close(kdt.Ch)
			// TODO - May need to remove this
			wg.Done()
			log.Print("Context closed while in Decrypt, closed channel.")
			return err
		} else {
			close(kdt.Ch)
			wg.Done()
			return err
		}
	}
	input := &kms.DecryptInput{
		CiphertextBlob: decodedBytes,
	}
	log.Print("Calling KMS Decryption Service...")
	response, err := KMS.DecryptWithContext(kdt.ctx, input)
	if err != nil && err.Error() == request.CanceledErrorCode {
		close(kdt.Ch)
		// TODO - May need to remove this
		wg.Done()
		log.Print("Context closed while in Decrypt, closed channel.")
		return err
	} else if err != nil {
		close(kdt.Ch)
		// TODO Same here
		wg.Done()
		return err
	}
	// Plaintext is a byte array, so convert to string
	//kdt.Ch <- string(response.Plaintext[:])
	//close(ch)
	log.Print("Finished A KMS Decyption Go Routine.")
	//}()
	select {
	case kdt.Ch <- string(response.Plaintext[:]) :
		log.Print("KMS Response Channel select")
		log.Print(string(response.Plaintext[:]))
		wg.Done()
		return nil
	case <-kdt.t.Dying():
		log.Print("KMS Tomb Dying... ")
		close(kdt.Ch)
		wg.Done()
		return nil
	}
}

func (kdt *KMSDecryptTomb) Stop() error {
	kdt.t.Kill(nil)
	return kdt.t.Wait()
}

func NewKMS() *kms.KMS {
	log.Print("Init KMS Config")
	c := kms.New(sess)
	xray.AWS(c.Client)
	return c
}

func NewBugCasterConfig(ctx context.Context) {
	log.Print("NewBugCasterConfig")
	KMS = NewKMS()
	cfg = new(BugCasterConfig)
	cfg.LoadConfig(ctx)
}

func NewKMSDecryptTomb(ctx context.Context, s string) *KMSDecryptTomb {
	kdt := &KMSDecryptTomb{
		ctx: ctx,
		s:   s,
		Ch:  make(chan string),
	}
	kdt.t.Go(kdt.decrypt)
	return kdt
}

func KMSDecrytiponWaiter() {
	log.Print("Waiting on KMS Decryption...")
	cfg.FishRatingUrl = <-tombFR.Ch
	log.Printf("FRU: %s", cfg.FishRatingUrl)
	cfg.GeoKey = <-tombGK.Ch
	log.Printf("GK: %s", cfg.GeoKey)
	wg.Wait()
	log.Print("Done Waiting On KMS Decryption.")
}

func init() {
	log.Print("Init Xray in Config")
	xray.Configure(xray.Config{
		LogLevel: "info",
	})
}

func (cfg *BugCasterConfig) LoadConfig(ctx context.Context) {
	log.Print("Begin LoadConfig")
	wg.Add(2)
	cfg.AlexaLocEndpoint = AlexaLocEndpoint
	tombFR = NewKMSDecryptTomb(ctx, os.Getenv("FISH_RATING_SERVICE_URL"))
	tombGK = NewKMSDecryptTomb(ctx, os.Getenv("GEO_KEY"))
	cfg.GeoUrl = os.Getenv("GEO_SERVICE_URL")
}
