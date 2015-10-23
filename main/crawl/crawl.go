package main

import (
	"fmt"

	"github.com/vikstrous/gotox/dht"
	"github.com/vikstrous/tox-crawler/crawler"
)

func main() {
	crawl, err := crawler.New()
	if err != nil {
		fmt.Printf("Failed to create server %s.\n", err)
		return
	}

	ch := make(chan struct{})
	go crawl.Listen(ch)

	for _, server := range dht.DhtServerList {
		err := crawl.Send(&dht.GetNodes{
			RequestedNodeID: &server.PublicKey,
		}, &server)
		if err != nil {
			fmt.Printf("error %s\n", err)
			return
		}
	}

	<-ch
	fmt.Printf("total: %d\n", len(crawl.AllPeers))
}
