package main

import (
	"ax-distiller/lib/extract"
	mdrender "ax-distiller/lib/md-render"
	"encoding/json"
	"log"
	"net/url"
	"os"
)

const SEARCH_TERM = "food wastage in restaurants and other corporate places"

func main() {
	os.Mkdir("results", 0777)

	extractor, err := extract.NewExtractor()
	if err != nil {
		log.Fatal(err)
	}
	parsed, err := url.Parse("https://en.wikipedia.org/wiki/Japan_Airlines")
	if err != nil {
		log.Fatal(err)
	}
	markdown, ax, err := extractor.Extract(parsed)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("results/markdown.md", []byte(mdrender.Render(markdown)), 0777)
	if err != nil {
		log.Fatal(err)
	}

	marshalled, err := json.Marshal(ax)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("results/axtree.json", []byte(marshalled), 0777)
	if err != nil {
		log.Fatal(err)
	}
}
