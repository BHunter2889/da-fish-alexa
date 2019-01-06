package apl

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ReadAplDocumentFromJsonFile(fileName string) (APLDocument, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return APLDocument{}, err
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	var apl APLDocument
	if err := json.Unmarshal(bytes, &apl); err != nil {
		return apl, err
	}

	return apl, nil
}

// `json:"document,omitempty"`
type APLDocument struct {
	Type         string `json:"type,omitempty"`
	Version      string `json:"version,omitempty"`
	Theme        string `json:"theme,omitempty"`
	MainTemplate struct {
		Description string   `json:"description,omitempty"`
		Parameters  []string `json:"parameters,omitempty"`
		Items       []struct {
			Type      string `json:"type,omitempty"`
			Direction string `json:"direction,omitempty"`
			Width     string `json:"width,omitempty"`
			Height    string `json:"height,omitempty"`
			Items     []struct {
				Type   string `json:"type,omitempty"`
				Source string `json:"source,omitempty"`
			} `json:"items,omitempty"`
		} `json:"items,omitempty"`
	} `json:"mainTemplate,omitempty"`
}