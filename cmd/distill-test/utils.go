package main

import (
	"ax-distiller/lib/ax"
	"encoding/json"
	"net/url"
	"os"
)

func fetchAxTree(link string) (ax.Node, error) {
	u, err := url.Parse(link)
	if err != nil {
		return ax.Node{}, err
	}

	navigator, err := ax.NewNavigator()
	if err != nil {
		return ax.Node{}, err
	}
	page, err := navigator.Navigate(u)
	if err != nil {
		return ax.Node{}, err
	}

	res, err := json.Marshal(page.Tree)
	if err != nil {
		return ax.Node{}, err
	}
	err = os.WriteFile("out.json", res, 0777)
	if err != nil {
		return ax.Node{}, err
	}

	return page.Tree, nil
}

func cachedAxTree(file string) (ax.Node, error) {
	buff, err := os.ReadFile(file)
	if err != nil {
		return ax.Node{}, err
	}

	node := ax.Node{}
	err = json.Unmarshal(buff, &node)
	if err != nil {
		return ax.Node{}, err
	}

	return node, nil
}
