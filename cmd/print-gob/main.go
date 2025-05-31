package main

import (
	"ax-distiller/lib/dnode"
	"encoding/gob"
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	flag.Parse()

	file := flag.Arg(0)

	fmt.Printf("reading from %s\n", file)

	f, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)

	var output *dnode.Node
	err = decoder.Decode(&output)
	if err != nil {
		panic(err)
	}

	spew.Dump(output)
}
