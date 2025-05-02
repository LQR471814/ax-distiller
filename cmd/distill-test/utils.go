package main

import (
	"ax-distiller/lib/ax"
	"encoding/json"
	"net/url"
	"os"
)

func fetchAxTree(link string) (ax.AXNode, error) {
	u, err := url.Parse(link)
	if err != nil {
		return ax.AXNode{}, err
	}

	navigator, err := ax.NewNavigator()
	if err != nil {
		return ax.AXNode{}, err
	}
	page, err := navigator.Navigate(u)
    if err != nil {
        return ax.AXNode{}, err
    }

	res, err := json.Marshal(page.Tree)
	if err != nil {
		return ax.AXNode{}, err
	}
	err = os.WriteFile("out.json", res, 0777)
	if err != nil {
		return ax.AXNode{}, err
	}

	return page.Tree, nil
}

func cachedAxTree(file string) (ax.AXNode, error) {
	buff, err := os.ReadFile(file)
	if err != nil {
		return ax.AXNode{}, err
	}

	ax := ax.AXNode{}
	err = json.Unmarshal(buff, &ax)
	if err != nil {
		return ax.AXNode{}, err
	}

	return ax, nil
}
