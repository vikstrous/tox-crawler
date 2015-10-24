// +build integration

package dht

import (
	"net"
	//"reflect"
	"testing"
	"time"
)

func TestBootstrap(t *testing.T) {
	dht, err := New()
	if err != nil {
		t.Fatalf("Failed to create server %s.", err)
	}
	go dht.Serve()
	defer dht.Stop()

	dht.AddFriend(&qToxPublicKey)

	node := Node{
		Addr: net.UDPAddr{
			IP:   net.ParseIP("::1"),
			Port: 33445,
		},
		PublicKey: qToxPublicKey,
	}
	dht.Bootstrap(node)

	//ping
	data, err := dht.PackPingPong(true, 1, &node.PublicKey)
	if err != nil {
		t.Errorf("error %s", err)
	}
	err = dht.Send(data, &node.Addr)
	if err != nil {
		t.Errorf("error %s", err)
	}

	// getnodes
	data, err = dht.PackGetNodes(&DhtServerList[0].PublicKey, qToxPublicKey)
	if err != nil {
		t.Errorf("error %s", err)
	}
	err = dht.Send(data, &DhtServerList[0].Addr)
	if err != nil {
		t.Errorf("error %s", err)
	}

	time.Sleep(time.Second)
	//<-dht.request
}
