package alexa

import (
	"encoding/json"
	"github.com/BHunter2889/da-fish-alexa/alexa/apl"
	"io/ioutil"
	"os"
)

const renderDirectiveType = "Alexa.Presentation.APL.RenderDocument"

type Directives struct {
	Directives []Directive
}

type DirectiveOption func(dir *Directives) (*Directives, error)

func (directives *Directives) NewBasicRenderDocumentDirective(token string, document apl.APLDocument, sources DataSources) {
	if len(directives.Directives) == 0 || &directives.Directives == nil {
		directives.Directives = make([]Directive, 1)
	}
	directives.Directives = append(directives.Directives, Directive{
		Type:        "Alexa.Presentation.APL.RenderDocument",
		Token:       token,
		Document:    document,
		DataSources: sources,
	})
}

func NewBasicAPLDirectives(token string, document apl.APLDocument, sources DataSources) Directives {
	d := Directives{}
	d.NewBasicRenderDocumentDirective(token, document, sources)
	return d
}

func ExtractNewRenderDocDirectiveFromJson(token string, fileName string, out *Directive) error {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}

	out.Token = token
	out.Type = renderDirectiveType

	return nil
}

//func (dir *Directives) BuildDirectives() {
//
//}
//
//func ExtractRenderDocDirectiveOption(token string, fileName string) DirectiveOption {
//	return func(dir *Directives) (directives *Directives, e error) {
//
//	}
//	jsonFile, err := os.Open(fileName)
//	if err != nil {
//		return err
//	}
//	defer jsonFile.Close()
//
//	bytes, _ := ioutil.ReadAll(jsonFile)
//
//	if err := json.Unmarshal(bytes, &out); err != nil {
//		return err
//	}
//
//	out.Token = token
//	out.Type = renderDirectiveType
//
//	return nil
//}

func NewDirectivesList(opts ...Directive) []Directive {
	dl := make([]Directive, 0)

	for _, opt := range opts {
		dl = append(dl, opt)
	}

	return dl
}
