package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/vikstrous/gotox"
	"github.com/vikstrous/gotox/dht"
)

type PeerInfo struct {
	dht.DHTPeer
	NumRequests int
}

// Crawler implements receive
type Crawler struct {
	dht.Transport
	// this holds all nodes discovered
	AllPeersMutex sync.Mutex
	AllPeers      map[[gotox.PublicKeySize]byte]PeerInfo
	Done          bool
}

func NewCrawler() (*Crawler, error) {
	id, err := dht.GenerateIdentity()
	if err != nil {
		return nil, err
	}
	transport, err := dht.NewUDPTransport(id)
	if err != nil {
		return nil, err
	}
	s := Crawler{
		Transport: transport,
		AllPeers:  make(map[[gotox.PublicKeySize]byte]PeerInfo),
	}
	transport.RegisterReceiver(s.Receive)

	go transport.Listen()

	go s.pingerTask()

	return &s, nil
}

func (s *Crawler) pingerTask() {
	for {
		// XXX: figure out the "right" interval for this
		numPeers := len(s.AllPeers)
		fmt.Printf("peers: %d\n", numPeers)
		duration := time.Duration(uint64(math.Log(float64(numPeers)))) * 200
		s.AllPeersMutex.Lock()
		done := true
		for _, neighbour := range s.AllPeers {
			// crawl only ipv4
			if neighbour.Addr.IP.To4() != nil {
				if neighbour.NumRequests < 10 {
					done = false
					err := s.Transport.Send(&dht.GetNodes{
						RequestedNodeID: &neighbour.PublicKey,
					}, &neighbour.DHTPeer)
					if err != nil {
						fmt.Println(err)
					}
					randomPK := [gotox.PublicKeySize]byte{}
					rand.Read(randomPK[:])
					err = s.Transport.Send(&dht.GetNodes{
						RequestedNodeID: &randomPK,
					}, &neighbour.DHTPeer)
					if err != nil {
						fmt.Println(err)
					}
					neighbour.NumRequests++
					s.AllPeers[neighbour.PublicKey] = neighbour
				}
			}
		}
		s.AllPeersMutex.Unlock()
		time.Sleep(duration * time.Millisecond)
		if numPeers == 0 {
			time.Sleep(time.Second)
		} else {
			if done {
				s.Done = true
				fmt.Println("done.")
				return
			}
		}
	}
}

func (s *Crawler) Receive(pp *dht.PlainPacket, addr *net.UDPAddr) bool {
	if s.Done {
		return true
	}
	switch payload := pp.Payload.(type) {
	case *dht.GetNodesReply:
		// There are only 4 replies
		s.AllPeersMutex.Lock()
		for _, node := range payload.Nodes {
			peer, found := s.AllPeers[node.PublicKey]
			// prefer ipv4
			if !found {
				s.AllPeers[node.PublicKey] = PeerInfo{DHTPeer: dht.DHTPeer{node.PublicKey, node.Addr}}
			} else {
				if peer.Addr.IP.To4() == nil && node.Addr.IP.To4() != nil {
					s.AllPeers[node.PublicKey] = PeerInfo{DHTPeer: dht.DHTPeer{node.PublicKey, node.Addr}}
				}
			}
		}
		s.AllPeersMutex.Unlock()
	default:
		fmt.Printf("Internal error. Failed to handle payload of parsed packet. %d", pp.Payload.Kind())
	}
	return false
}

func main() {
	crawler, err := NewCrawler()
	if err != nil {
		fmt.Printf("Failed to create server %s.\n", err)
		return
	}

	go crawler.Listen()

	for _, server := range dht.DhtServerList[:5] {
		err := crawler.Send(&dht.GetNodes{
			RequestedNodeID: &server.PublicKey,
		}, &server)
		if err != nil {
			fmt.Printf("error %s\n", err)
			return
		}
	}

	time.Sleep(time.Hour * 1000000)
}
