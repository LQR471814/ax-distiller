package main

import (
	"archive/zip"
	"ax-distiller/lib/utils"
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/bitfield/script"
)

func main() {
	os.Mkdir("data", 0777)

	if _, err := os.Stat("data/chrome"); os.IsNotExist(err) {
		log.Println("downloading chrome...")
		_, err := script.
			Get("https://github.com/ungoogled-software/ungoogled-chromium-portablelinux/releases/download/136.0.7103.113-1/ungoogled-chromium_136.0.7103.113-1_linux.tar.xz").
			WriteFile("data/chrome.tar.xz")
		if err != nil {
			log.Fatal(err)
		}

		err = os.MkdirAll("data/chrome", 0777)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("extracting chrome...")
		p := script.Exec("tar -xf data/chrome.tar.xz -C data/chrome --strip-components=1")
		p.Wait()

		err = os.Remove("data/chrome.tar.xz")
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
	}

	assetsOriginal, err := os.ReadFile("data/ublock/assets/assets.json")
	if err != nil {
		log.Fatal(err)
	}

	var assets map[string]map[string]any
	err = json.Unmarshal(assetsOriginal, &assets)
	if err != nil {
		log.Fatal(err)
	}

	addFilter := func(id, title, url string) {
		assets[id] = map[string]any{
			"content":    "filters",
			"group":      "annoyances",
			"off":        false,
			"title":      title,
			"contentURL": url,
			"supportURL": url,
			"cdnURLs": []string{
				url,
			},
		}
	}

	addFilter(
		"custom-block-remote-fonts",
		"Block Third-Party Fonts",
		"https://raw.githubusercontent.com/yokoffing/filterlists/refs/heads/main/block_third_party_fonts.txt",
	)
	addFilter(
		"custom-browse-websites-without-logging-in",
		"Browse Websites Without Logging In",
		"https://raw.githubusercontent.com/DandelionSprout/adfilt/refs/heads/master/BrowseWebsitesWithoutLoggingIn.txt",
	)
	addFilter(
		"custom-adblock-pro",
		"Adblock Pro",
		"https://raw.githubusercontent.com/hagezi/dns-blocklists/refs/heads/main/adblock/pro.mini.txt",
	)
	addFilter(
		"custom-spam-tlds-ublock",
		"Spam Blocker",
		"https://raw.githubusercontent.com/hagezi/dns-blocklists/refs/heads/main/adblock/spam-tlds-ublock.txt",
	)
	addFilter(
		"custom-ublock-combo",
		"uBlock Combo",
		"https://raw.githubusercontent.com/iam-py-test/uBlock-combo/refs/heads/main/list.txt",
	)
	addFilter(
		"custom-annoyance-list",
		"Annoyance List",
		"https://raw.githubusercontent.com/yokoffing/filterlists/refs/heads/main/annoyance_list.txt",
	)
	addFilter(
		"custom-click2load",
		"Click2Load",
		"https://raw.githubusercontent.com/yokoffing/filterlists/refs/heads/main/click2load.txt",
	)
	addFilter(
		"custom-privacy-essentials",
		"Privacy Essentials",
		"https://raw.githubusercontent.com/yokoffing/filterlists/refs/heads/main/privacy_essentials.txt",
	)

	assets["easylist-annoyances"]["off"] = false
	assets["easylist-chat"]["off"] = false
	assets["easylist-newsletters"]["off"] = false
	assets["easylist-notifications"]["off"] = false

	modified, err := json.MarshalIndent(assets, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("data/ublock/assets/assets.json", modified, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("setup complete.")
}
