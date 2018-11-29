package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"github.com/BHunter2889/da-fish-alexa/services"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-xray-sdk-go/xray"
	"gopkg.in/tomb.v2"
	"log"
	"os"
	"strings"
)

func killAllTombs(ctx context.Context) {
	if tombFR != nil {
		addAndHandleXRayRecordingError(ctx, tombFR.Stop())
	}
	if tombGK != nil {
		addAndHandleXRayRecordingError(ctx, tombGK.Stop())
	}
	if tombLoc != nil {
		addAndHandleXRayRecordingError(ctx, tombLoc.Stop())
	}
}

// Device Locator Tomb
type DeviceLocator struct {
	ctx              context.Context
	DeviceLocService *services.DeviceService
	Ch               chan *alexa.DeviceLocationResponse
	ErCh             chan error
	t                tomb.Tomb
}

func NewDeviceLocator(ctx context.Context, ds *services.DeviceService) *DeviceLocator {
	dl := &DeviceLocator{
		ctx:              ctx,
		DeviceLocService: ds,
		Ch:               make(chan *alexa.DeviceLocationResponse),
		ErCh:             make(chan error),
	}
	dl.t.Go(dl.getDeviceLocation)
	return dl
}

func (dl *DeviceLocator) Stop() error {
	dl.t.Kill(nil)
	return dl.t.Wait()
}

func (dl *DeviceLocator) getDeviceLocation() error {
	resp, err := deviceLocService.GetDeviceLocation(dl.ctx)
	if err != nil {
		log.Print(resp)
		log.Print(err)
		if strings.Contains(err.Error(), "403") {
			dl.ErCh <- err
			close(dl.Ch)
			close(dl.ErCh)
			return nil
		}
		close(dl.Ch)
		close(dl.ErCh)
		return err
	}

	select {
	case dl.Ch <- resp:
		close(dl.ErCh)
		return nil
	case <-dl.t.Dying():
		log.Print("DeviceLocator Dying...")
		close(dl.Ch)
		close(dl.ErCh)
		return nil
	}
}

// KMS Decryption Tomb
type KMSDecryptTomb struct {
	ctx  context.Context
	name string
	s    string
	Ch   chan string
	t    tomb.Tomb
}

func NewKMSDecryptTomb(ctx context.Context, s string) *KMSDecryptTomb {
	kdt := &KMSDecryptTomb{
		ctx:  ctx,
		name: s,
		s:    os.Getenv(s),
		Ch:   make(chan string),
	}
	kdt.t.Go(kdt.decrypt)
	return kdt
}

func (kdt *KMSDecryptTomb) Stop() error {
	kdt.t.Kill(nil)
	return kdt.t.Wait()
}

// We want this in a channel
// Logging For Demo Purposes
// seg - the name for the x-ray tracing subsegment, should represent the env var being decrypted.
func (kdt *KMSDecryptTomb) decrypt() (ctxErr error) {
	segName := fmt.Sprintf("KMSDecrypt.%s", kdt.name)

	// This allows us to capture the performance of the entire decryption process.
	// Do Not use `xray.CaptureAsync` even though per docs it seems appropriate. Breaks Everything.
	err := xray.Capture(kdt.ctx, segName, func(ctx1 context.Context) error {
		//defer wg.Done()

		log.Print("New Decrypt...")
		decodedBytes, err := base64.StdEncoding.DecodeString(kdt.s)
		if err != nil {
			log.Print(err)

			// Conditional Exists here solely for Demoing Context Cancellation.
			if err.Error() == request.CanceledErrorCode {
				close(kdt.Ch)
				wg.Done()
				log.Print("Context closed while in Decrypt, closed channel.")
				ctxErr = err
				return err
			} else {
				close(kdt.Ch)
				wg.Done()
				ctxErr = err
				return err
			}
		}
		input := &kms.DecryptInput{
			CiphertextBlob: decodedBytes,
		}
		log.Print("Calling kMS Decryption Service...")
		response, err := kMS.DecryptWithContext(kdt.ctx, input)

		// Conditional Exists here solely for Demoing Context Cancellation.
		if err != nil && err.Error() == request.CanceledErrorCode {
			close(kdt.Ch)
			log.Print("Context closed while in Decrypt, closed channel.")
			addAndHandleXRayRecordingError(ctx1, err)
			wg.Done()
			ctxErr = err
			return err
		} else if err != nil {
			close(kdt.Ch)
			addAndHandleXRayRecordingError(ctx1, err)
			wg.Done()
			ctxErr = err
			return err
		}
		log.Print("Finished A kMS Decyption Go Routine.")

		// Listen for either successful decryption or a Context Cancellation related event.
		// Plaintext is a byte array, so convert to string
		select {
		case kdt.Ch <- string(response.Plaintext[:]):
			log.Print("kMS Response Channel select")
			log.Print(string(response.Plaintext[:]))
			wg.Done()
			return nil
		case <-kdt.t.Dying():
			log.Print("kMS Tomb Dying... ")
			close(kdt.Ch)
			addAndHandleXRayRecordingError(ctx1, xray.AddMetadata(ctx1, segName, "KMS Tomb is Dying"))
			wg.Done()
			return nil
		}
	})
	addAndHandleXRayRecordingError(kdt.ctx, err)
	return
}
