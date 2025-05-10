package main

import (
	"ax-distiller/lib/ax"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/url"
)

func main() {
	flag.Parse()
	u := flag.Arg(0)

	if u == "" {
		log.Fatal("you must provide a url to dump as the first command line argument")
	}

	navigator, err := ax.NewNavigator(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		log.Fatal(err)
	}

	page, err := navigator.Navigate(parsed)
	if err != nil {
		log.Fatal(err)
	}

	root := page.Tree
	allWhitespace := ax.FilterWhitespace(&root)
	if allWhitespace {
		log.Println("there is nothing except whitespace in the resulting website, therefore there is no output")
		return
	}

	serialized, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(serialized))
}
