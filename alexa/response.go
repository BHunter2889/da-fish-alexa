package alexa

import "da-fish-alexa/alexa"

func NewSimpleResponse(title string, text string) Response {
	r := Response{
		Version: "1.0",
		Body: ResBody{
			OutputSpeech: &Payload{
				Type: "PlainText",
				Text: text,
			},
			Card: &Payload{
				Type:    "Simple",
				Title:   title,
				Content: text,
			},
			ShouldEndSession: true,
		},
	}
	return r
}

// Consider adding custom prompt if possible
func NewPermissionsRequestResponse() Response {
	var builder alexa.SSMLBuilder
	builder.Say("Bug Caster was unable to access your device's zip code and country information. ")
	builder.Pause("750")
	builder.Say("If you have not enabled Bug Caster to access this information, ")
	builder.Pause("150")
	builder.Say("Please check your Alexa App to grant permission for Bug Caster to access your zip code and country " +
		"information so that the fishing forecast for your area may be determined. ")
	r := Response{
		Version: "1.0",
		Body: ResBody{
			OutputSpeech: &Payload{
				Type: "PlainText",
				Text: builder.Build(),
			},
			Card: &Payload{
				Type:        "AskForPermissionsConsent",
				Permissions: []string{"read::alexa:device:all:address:country_and_postal_code"},
			},
		},
	}
	return r
}

func NewDefaultErrorResponse() Response {
	var builder alexa.SSMLBuilder
	builder.Say("Bug Caster is down and undergoing maintenance. ")
	builder.Pause("750")
	builder.Say(" Apologies for the inconvenience. ")
	builder.Pause("500")
	builder.Say("Please try Bug Caster again later.")

	r := Response{
		Version: "1.0",
		Body: ResBody{
			OutputSpeech: &Payload{
				Type: "PlainText",
				Text: builder.Build(),
			},
			Card: &Payload{
				Type:  "Simple",
				Title: "BugCaster Under Maintenance",
				Text:  builder.Build(),
			},
		},
	}
	return r
}

type Response struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Body              ResBody                `json:"response"`
}

type ResBody struct {
	OutputSpeech     *Payload     `json:"outputSpeech,omitempty"`
	Card             *Payload     `json:"card,omitempty"`
	Reprompt         *Reprompt    `json:"reprompt,omitempty"`
	Directives       []Directives `json:"directives,omitempty"`
	ShouldEndSession bool         `json:"shouldEndSession"`
}

type Reprompt struct {
	OutputSpeech Payload `json:"outputSpeech,omitempty"`
}

type Directives struct {
	Type          string         `json:"type,omitempty"`
	SlotToElicit  string         `json:"slotToElicit,omitempty"`
	UpdatedIntent *UpdatedIntent `json:"UpdatedIntent,omitempty"`
	PlayBehavior  string         `json:"playBehavior,omitempty"`
	AudioItem     struct {
		Stream struct {
			Token                string `json:"token,omitempty"`
			URL                  string `json:"url,omitempty"`
			OffsetInMilliseconds int    `json:"offsetInMilliseconds,omitempty"`
		} `json:"stream,omitempty"`
	} `json:"audioItem,omitempty"`
}

type UpdatedIntent struct {
	Name               string                 `json:"name,omitempty"`
	ConfirmationStatus string                 `json:"confirmationStatus,omitempty"`
	Slots              map[string]interface{} `json:"slots,omitempty"`
}

type Image struct {
	SmallImageURL string `json:"smallImageUrl,omitempty"`
	LargeImageURL string `json:"largeImageUrl,omitempty"`
}

type Payload struct {
	Type        string   `json:"type,omitempty"`
	Title       string   `json:"title,omitempty"`
	Text        string   `json:"text,omitempty"`
	SSML        string   `json:"ssml,omitempty"`
	Content     string   `json:"content,omitempty"`
	Image       Image    `json:"image,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// Response(s) from requests made back to the Alexa Api

type DeviceLocationResponse struct {
	CountryCode string `json:"countryCode"`
	PostalCode  string `json:"postalCode"`
}
