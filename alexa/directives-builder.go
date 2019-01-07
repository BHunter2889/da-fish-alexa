package alexa

import (
	"github.com/BHunter2889/da-fish-alexa/alexa/apl"
)

type Directives struct {
	Directives []Directive
}

func (directives *Directives) NewBasicRenderDocumentDirective(token string, document apl.APLDocument, sources DataSources) {
	if len(directives.Directives) == 0 || &directives.Directives == nil {
		directives.Directives = make([]Directive, 1)
	}
	directives.Directives = append(directives.Directives, Directive{
		Type: "Alexa.Presentation.APL.RenderDocument",
		Token: token,
		Document: document,
		DataSources: sources,
	})
}

func NewBasicAPLDirectives(token string, document apl.APLDocument, sources DataSources) Directives {
	d := Directives{}
	d.NewBasicRenderDocumentDirective(token, document, sources)
	return d
}