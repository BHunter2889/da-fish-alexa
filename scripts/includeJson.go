package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

// Reads the bugcaster-apl-doc-2.json file in the current folder
// and encodes it as a string literal in apl-json.go
// While it would be faster to read the file itself instead of the directory and then
// iterate over it, this is done at compile time so it hardly matters and this can later serve
// to read a config.json file into the code as well, for example.
func main() {
	fs, _ := ioutil.ReadDir(".")
	out, _ := os.Create("apl-json.go")
	if _, err := out.Write([]byte("package main \n\nconst (\n")); err != nil {
		log.Println(err)
	}

	for _, f := range fs {
		if f.Name() == "bugcaster-apl-doc-2.json" {
			if _, err := out.Write([]byte("aplJson = `")); err != nil {
				log.Println(err)
			}
			f, _ := os.Open(f.Name()) // Open file
			// Copy contents
			if _, err := io.Copy(out, f); err != nil {
				log.Println(err)
			}
			// Assign contents as string literal to const aplJson
			if _, err := out.Write([]byte("`\n")); err != nil {
				log.Println(err)
			}
		}
	}
	// Write a line and finish
	if _, err := out.Write([]byte(")\n")); err != nil {
		log.Println(err)
	}
}
