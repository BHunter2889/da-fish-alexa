package alexa

// Credit - Nic Raboy: Modified version of Arien Malec's work
// https://github.com/nraboy/alexa-slick-dealer/blob/master/alexa/request.go
// https://github.com/arienmalec/alexa-go
// https://medium.com/@amalec/alexa-skills-with-go-54db0c21e758

const (
	HelpIntent   = "AMAZON.HelpIntent"
	CancelIntent = "AMAZON.CancelIntent"
	StopIntent   = "AMAZON.StopIntent"
)

type Request struct {
	Version string  `json:"version"`
	Session Session `json:"session"`
	Body    ReqBody `json:"request"`
	Context Context `json:"context"`
}

type Session struct {
	New         bool   `json:"new"`
	SessionID   string `json:"sessionId"`
	Application struct {
		ApplicationID string `json:"applicationId"`
	} `json:"application"`
	Attributes map[string]interface{} `json:"attributes"`
	User       struct {
		UserID      string `json:"userId"`
		AccessToken string `json:"accessToken,omitempty"`
	} `json:"user"`
}

type Context struct {
	System struct {
		APIAccessToken string `json:"apiAccessToken"`
		APIEndpoint    string `json:"apiEndpoint"`
		Device         struct {
			DeviceID            string              `json:"deviceId,omitempty"`
			SupportedInterfaces SupportedInterfaces `json:"supportedInterfaces,omitempty"`
		} `json:"device,omitempty"`
		Application struct {
			ApplicationID string `json:"applicationId,omitempty"`
		} `json:"application,omitempty"`
	} `json:"System,omitempty"`
}

// Interfaces Supported by the User's device. This is not comprehensive.
type SupportedInterfaces struct {
	APL struct {
		Runtime struct {
			MaxVersion string `json:"maxVersion,omitempty"`
		} `json:"runtime,omitempty"`
	} `json:"Alexa.Presentation.APL,omitempty"`
	AudioPlayer struct{} `json:"AudioPlayer,omitempty"` // This appears to always be an empty object
}
/**
APL Document UserEvents
see: https://developer.amazon.com/docs/alexa-presentation-language/apl-support-for-your-skill.html#listen-for-apl-userevents-from-alexa

Usage: `json:"event,omitempty"`
 */
type Event struct {
	Source struct {
		Type    string      `json:"type,omitempty"`
		Handler string      `json:"handler,omitempty"`
		ID      string      `json:"id,omitempty"`
		Value   interface{} `json:"value,omitempty"`
	} `json:"source,omitempty"`
	Arguments []string `json:"arguments,omitempty"`
}

type ReqBody struct {
	Type        string `json:"type"`
	RequestID   string `json:"requestId"`
	Timestamp   string `json:"timestamp"`
	Locale      string `json:"locale"`
	Token       string `json:"token,omitempty"`
	Event       Event  `json:"event,omitempty"`
	Intent      Intent `json:"intent,omitempty"`
	Reason      string `json:"reason,omitempty"`
	DialogState string `json:"dialogState,omitempty"`
}

type Intent struct {
	Name  string          `json:"name"`
	Slots map[string]Slot `json:"slots"`
}

type Slot struct {
	Name        string      `json:"name"`
	Value       string      `json:"value"`
	Resolutions Resolutions `json:"resolutions"`
}

type Resolutions struct {
	ResolutionPerAuthority []struct {
		Values []struct {
			Value struct {
				Name string `json:"name"`
				Id   string `json:"id"`
			} `json:"value"`
		} `json:"values"`
	} `json:"resolutionsPerAuthority"`
}
