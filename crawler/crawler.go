package crawler

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
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
	doneChan      chan struct{}
}

func New() (*Crawler, error) {
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
		doneChan:  make(chan struct{}),
	}

	go s.Receive()

	go s.pingerTask()

	return &s, nil
}

func (s *Crawler) Wait() {
	<-s.doneChan
}

func (s *Crawler) Crawl() []dht.DHTPeer {
	go s.Listen()
	s.Wait()
	peers := []dht.DHTPeer{}
	for _, peerInfo := range s.AllPeers {
		peers = append(peers, peerInfo.DHTPeer)
	}
	return peers
}

func (s *Crawler) pingerTask() {
	for {
		// XXX: figure out the "right" interval for this
		numPeers := len(s.AllPeers)
		duration := time.Duration(uint64(math.Log(float64(numPeers)))) * 200
		s.AllPeersMutex.Lock()
		done := true
		for _, neighbour := range s.AllPeers {
			// crawl only ipv4
			if neighbour.Addr.IP.To4() != nil && neighbour.NumRequests < 10 {
				done = false
				err := s.Transport.Send(&dht.GetNodes{
					RequestedNodeID: &neighbour.PublicKey,
				}, &neighbour.DHTPeer)
				if err != nil {
					log.Println(err)
				}
				randomPK := [gotox.PublicKeySize]byte{}
				rand.Read(randomPK[:])
				err = s.Transport.Send(&dht.GetNodes{
					RequestedNodeID: &randomPK,
				}, &neighbour.DHTPeer)
				if err != nil {
					log.Println(err)
				}
				neighbour.NumRequests++
				s.AllPeers[neighbour.PublicKey] = neighbour
			}
		}
		s.AllPeersMutex.Unlock()
		time.Sleep(duration * time.Millisecond)
		if numPeers == 0 {
			for _, server := range dht.DhtServerList {
				err := s.Send(&dht.GetNodes{
					RequestedNodeID: &server.PublicKey,
				}, &server)
				if err != nil {
					fmt.Printf("error contacting bootstap node %v %s\n", server, err)
				}
			}
			time.Sleep(time.Second)
		} else {
			if done {
				s.Transport.Stop()
				close(s.doneChan)
				return
			}
		}
	}
}

func (s *Crawler) Receive() {
	for {
		message, ok := <-s.Transport.DataChan()
		if !ok {
			return
		}
		switch payload := message.Packet.Payload.(type) {
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
			//fmt.Printf("Internal error. Failed to handle payload of parsed packet. %d\n", pp.Payload.Kind())
		}
	}
}
