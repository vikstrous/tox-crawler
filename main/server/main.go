package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"github.com/vikstrous/gotox/dht"
	"github.com/vikstrous/tox-crawler/crawler"
	"golang.org/x/net/websocket"
)

const DataFile = "data/data.json"
const CountryCountsFile = "data/peers.json"
const GeoIPFile = "./GeoIP.dat"
const MaxDatapoints = 14000

type Result struct {
	Time   time.Time `json:"time"`
	Number int       `json:"number"`
}

type Stats struct {
	NumbersList   []Result       `json:"numbers_list"`
	CountryCounts map[string]int `json:"country_counts"`
}

type Crawler struct {
	Mut       sync.Mutex
	Results   []Result
	ChMut     sync.Mutex
	ChResults []chan struct{}
	GeoIP     *maxminddb.Reader
	Counts    map[string]int
}

func NewCrawler() *Crawler {
	return &Crawler{}
}

func (c *Crawler) OneCrawl() []dht.DHTPeer {
	crawl, err := crawler.New()
	if err != nil {
		fmt.Printf("Failed to create server %s.\n", err)
		return nil
	}

	return crawl.Crawl()
}

type mmCountry struct {
	Country struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
}

func (c *Crawler) SaveState() {
	content, err := json.Marshal(c.Results)
	if err != nil {
		log.Printf("Error building json: %s", err)
	}
	log.Printf("Saving data...")
	err = ioutil.WriteFile(DataFile, content, 0644)
	if err != nil {
		log.Printf("Error writing data: %s", err)
	}
	log.Printf("Saved data...")
	content, err = json.Marshal(c.Counts)
	if err != nil {
		log.Printf("Error building json: %s", err)
	}
	log.Printf("Saving country counts...")
	err = ioutil.WriteFile(CountryCountsFile, content, 0644)
	if err != nil {
		log.Printf("Error writing data: %s", err)
	}
	log.Printf("Saved country counts...")
}

func (c *Crawler) Crawl() {
	for {
		log.Printf("Crawling...")
		peers := c.OneCrawl()
		numNodes := len(peers)
		log.Printf("Found %d nodes", numNodes)

		// if we have a geoip database, use it
		counts := map[string]int{}
		if c.GeoIP != nil {
			for _, peer := range peers {
				var record mmCountry
				err := c.GeoIP.Lookup(peer.Addr.IP, &record)
				if err != nil {
					log.Printf("Error reading from geoip database: %s", err)
				}
				counts[record.Country.Names["en"]] += 1
			}
		}

		c.Mut.Lock()
		// update graph
		c.Results = append(c.Results, Result{time.Now(), numNodes})
		if len(c.Results) > MaxDatapoints {
			c.Results = c.Results[1:]
		}
		c.Counts = counts

		c.SaveState()
		c.Mut.Unlock()

		// notify listeners that we added a new entry
		c.ChMut.Lock()
		log.Printf("Notifying %d websocket clients", len(c.ChResults))
		for _, ch := range c.ChResults {
			close(ch)
		}
		c.ChResults = [](chan struct{}){}
		c.ChMut.Unlock()

		// slow down our crawling interval to avoid using too much bandwidth
		log.Printf("Taking a nap until next interval")
		time.Sleep(5 * time.Minute)
	}
}

func (c *Crawler) statsHandler(ws *websocket.Conn) {
	c.Mut.Lock()
	json.NewEncoder(ws).Encode(Stats{c.Results, c.Counts})
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
			json.NewEncoder(ws).Encode(Stats{c.Results, c.Counts})
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
				json.NewEncoder(ws).Encode(Stats{c.Results, c.Counts})
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
	content, err := ioutil.ReadFile(CountryCountsFile)
	if err != nil {
		c.Counts = map[string]int{}
	} else {
		err = json.Unmarshal(content, &c.Counts)
		if err != nil {
			panic("Failed to load country stats: " + err.Error())
		} else {
			log.Printf("Loaded info on %d countries", len(c.Counts))
		}
	}
	content, err = ioutil.ReadFile(DataFile)
	if err != nil {
		c.Results = make([]Result, 0, MaxDatapoints)
	} else {
		err = json.Unmarshal(content, &c.Results)
		if err != nil {
			panic("Failed to load data: " + err.Error())
		} else {
			log.Printf("Loaded %d results", len(c.Results))
		}
	}

	gi, err := maxminddb.Open("GeoIP.dat")
	if err != nil {
		fmt.Printf("Could not open GeoIP database. Running in mapless mode.\n")
	}
	defer gi.Close()
	c.GeoIP = gi

	go c.Crawl()

	http.Handle("/stats", websocket.Handler(c.statsHandler))
	http.Handle("/", http.FileServer(http.Dir("server")))
	err = http.ListenAndServe(":7071", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
