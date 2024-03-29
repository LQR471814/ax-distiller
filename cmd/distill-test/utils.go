package main

import (
	"ax-distiller/lib/axextract"
	"encoding/json"
	"net/url"
	"os"
)

func fetchAxTree(link string) (axextract.AXNode, error) {
	u, err := url.Parse(link)
	if err != nil {
		return axextract.AXNode{}, err
	}

	navigator, err := axextract.NewNavigator()
	if err != nil {
		return axextract.AXNode{}, err
	}
	page, err := navigator.Navigate(u)
    if err != nil {
        return axextract.AXNode{}, err
    }

	res, err := json.Marshal(page.Tree())
	if err != nil {
		return axextract.AXNode{}, err
	}
	err = os.WriteFile("out.json", res, 0777)
	if err != nil {
		return axextract.AXNode{}, err
	}

	return page.Tree(), nil
}

func fetchAxPage(link string) (axextract.Page, error) {
	u, err := url.Parse(link)
	if err != nil {
		return axextract.Page{}, err
	}

	navigator, err := axextract.NewNavigator()
	if err != nil {
		return axextract.Page{}, err
	}

	return navigator.Navigate(u)
}

func cachedAxTree(file string) (axextract.AXNode, error) {
	buff, err := os.ReadFile(file)
	if err != nil {
		return axextract.AXNode{}, err
	}

	ax := axextract.AXNode{}
	err = json.Unmarshal(buff, &ax)
	if err != nil {
		return axextract.AXNode{}, err
	}

	return ax, nil
}
