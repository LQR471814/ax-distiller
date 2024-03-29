package main

import (
	"ax-distiller/lib/axextract"
	"log"
	"os"
)

func recurse(node axextract.AXNode, m map[string]struct{}) {
	for _, c := range node.Role {
		if c >= 65 && c <= 91 {
			m[node.Role] = struct{}{}
			break
		}
	}
	for _, c := range node.Children {
		recurse(c, m)
		os.Stdout.Write([]byte(c.Name))
	}
}

func handleTree() {
	if len(os.Args) != 2 {
		log.Fatalf("you must specify the file to import from")
	}

	ax, err := cachedAxTree(os.Args[1])
	// ax, err := fetchAxTree("https://en.wikipedia.org/wiki/Quantum_mechanics")
	if err != nil {
		log.Fatal(err)
	}

	out := map[string]struct{}{}
	recurse(ax, out)
	for k := range out {
		os.Stderr.Write([]byte(k + "\n"))
	}
}

func debugPage() {
	// page, err := fetchAxPage("https://en.wikipedia.org/wiki/Quantum_mechanics")
	page, err := fetchAxPage("https://www.npr.org/2024/03/29/1198909601/lost-animals-moles-rats-being-rediscovered")
	if err != nil {
		log.Fatal(err)
	}
	err = page.ShowDebugInfo()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	debugPage()
}
