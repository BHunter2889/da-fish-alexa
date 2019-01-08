package main

import (
	"github.com/BHunter2889/da-fish-alexa/alexa"
	"log"
)

const filename = "bugcaster-apl-doc-2.json"

func main() {
	testDirective := alexa.Directive{}
	if err := alexa.ExtractNewRenderDocDirectiveFromJson("testing", filename, &testDirective);
	err != nil {
		log.Print(err)
	}
	dl := alexa.NewDirectivesList(testDirective)
	log.Print("************HERE************", testDirective)
	log.Println(len(dl))
}
