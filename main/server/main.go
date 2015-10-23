package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/vikstrous/gotox/dht"
	"github.com/vikstrous/tox-crawler/crawler"
	"golang.org/x/net/websocket"
)

type Result struct {
	Time   time.Time `json:"time"`
	Number int       `json:"number"`
}

type Crawler struct {
	Mut       sync.Mutex
	Results   []Result
	ChMut     sync.Mutex
	ChResults []chan struct{}
}

func NewCrawler() *Crawler {
	return &Crawler{}
}

func (c *Crawler) OneCrawl() int {
	crawl, err := crawler.New()
	if err != nil {
		fmt.Printf("Failed to create server %s.\n", err)
		return 0
	}

	ch := make(chan struct{})
	go crawl.Listen(ch)

	for _, server := range dht.DhtServerList {
		err := crawl.Send(&dht.GetNodes{
			RequestedNodeID: &server.PublicKey,
		}, &server)
		if err != nil {
			fmt.Printf("error %s\n", err)
			return 0
		}
	}

	<-ch
	return len(crawl.AllPeers)
}
func (c *Crawler) Crawl() {
	for {
		numNodes := c.OneCrawl()
		c.Mut.Lock()
		c.Results = append(c.Results, Result{time.Now(), numNodes})
		if len(c.Results) > 20 {
			c.Results = c.Results[1:]
		}
		c.Mut.Unlock()
		c.ChMut.Lock()
		for _, ch := range c.ChResults {
			close(ch)
		}
		c.ChResults = [](chan struct{}){}
		c.ChMut.Unlock()
	}
}

func (c *Crawler) statsHandler(ws *websocket.Conn) {
	c.Mut.Lock()
	json.NewEncoder(ws).Encode(c.Results)
	c.Mut.Unlock()

	// if a message we receive exceeds 2000 bytes, it'll be broken up and could
	// trigger more than one reply
	buf := make([]byte, 2000)

	ch := make(chan struct{})
	c.ChMut.Lock()
	c.ChResults = append(c.ChResults, ch)
	c.ChMut.Unlock()

	for {
		select {
		case <-ch:
			// if we have data, send it
			c.Mut.Lock()
			json.NewEncoder(ws).Encode(c.Results)
			c.Mut.Unlock()

			ch = make(chan struct{})
			c.ChMut.Lock()
			c.ChResults = append(c.ChResults, ch)
			c.ChMut.Unlock()
		default:
			// else, keep waiting; process messages from the user
			ws.SetReadDeadline(time.Now().Add(time.Second))
			// if we receive any data,
			_, err := ws.Read(buf)
			if err == nil {
				// if we have data, send it
				c.Mut.Lock()
				json.NewEncoder(ws).Encode(c.Results)
				c.Mut.Unlock()
			} else if opErr, ok := err.(*net.OpError); !ok || !opErr.Timeout() {
				// non-timeout network error -> close the connection
				ws.Close()
				return
			}
		}
	}
}

func main() {

	c := NewCrawler()
	go c.Crawl()

	http.Handle("/stats", websocket.Handler(c.statsHandler))
	http.Handle("/", http.FileServer(http.Dir("server")))
	err := http.ListenAndServe(":7070", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
