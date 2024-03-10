package main

import (
	"encoding/xml"
	"log"
	"os"
)

type RDF struct {
	XMLName xml.Name `xml:"RDF"`
	Classes []Class  `xml:"Class"`
}

type Class struct {
	XMLName xml.Name     `xml:"Class"`
	Id      string       `xml:"ID,attr"`
	Extends []Superclass `xml:"subClassOf"`
}

type Superclass struct {
	XMLName xml.Name `xml:"subClassOf"`
	Name    string   `xml:"resource,attr"`
}

func main() {
	buff, err := os.ReadFile("./aria-1.rdf")
	if err != nil {
		log.Fatal(err)
	}

	rdf := RDF{}
	err = xml.Unmarshal(buff, &rdf)
	if err != nil {
		log.Fatal(err)
	}

	for _, class := range rdf.Classes {
		log.Println(class.Id, "-----")
		for _, superClass := range class.Extends {
			if superClass.Name == "" {
				continue
			}
			cropped := superClass.Name[1:]
			log.Println("\t" + cropped)
		}
	}
}
