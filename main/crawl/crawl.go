package main

import (
	"fmt"

	"github.com/vikstrous/tox-crawler/crawler"
)

func main() {
	crawl, err := crawler.New()
	if err != nil {
		fmt.Printf("Failed to create server %s.\n", err)
		return
	}

	peers := crawl.Crawl()

	fmt.Printf("total: %d\n", len(peers))
}
