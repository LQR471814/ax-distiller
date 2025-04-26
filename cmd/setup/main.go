package main

import (
	"archive/zip"
	"ax-distiller/lib/utils"
	"bytes"
	"log"
	"os"

	"github.com/bitfield/script"
)

func main() {
	os.Mkdir("data", 0777)

	if _, err := os.Stat("data/thorium-browser"); os.IsNotExist(err) {
		log.Println("downloading thorium...")
		_, err := script.
			Get("https://github.com/Alex313031/thorium/releases/download/M130.0.6723.174/Thorium_Browser_130.0.6723.174_AVX2.AppImage").
			WriteFile("data/thorium-browser")
		if err != nil {
			log.Fatal(err)
		}
		err = os.Chmod("data/thorium-browser", 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat("data/ublock"); os.IsNotExist(err) {
		log.Println("downloading ublock origin...")
		ublockArchive, err := script.
			Get("https://github.com/gorhill/uBlock/releases/download/1.63.2/uBlock0_1.63.2.chromium.zip").
			Bytes()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("extracting ublock origin...")
		unzippedUblock, err := zip.NewReader(bytes.NewReader(ublockArchive), int64(len(ublockArchive)))
		if err != nil {
			log.Fatal(err)
		}
		err = utils.Unzip(unzippedUblock, "data/ublock-temp")
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat("data/ublock-temp/uBlock.chromium"); os.IsNotExist(err) {
		err = os.Rename("data/ublock-temp/uBlock0.chromium", "data/ublock")
		if err != nil {
			log.Fatal(err)
		}
		err = os.Remove("data/ublock-temp")
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("setup complete.")
}
