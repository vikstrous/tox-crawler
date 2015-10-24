package main

import (
	"encoding/hex"
	"fmt"

	"github.com/vikstrous/gotox"
	"github.com/vikstrous/gotox/dht"
)

var publicKey = [gotox.PublicKeySize]byte{}

func init() {
	publicKeySlice, _ := hex.DecodeString("A4D28D52D4116A02147ECE6C6299DA3F5524DEBA043B067CF7D5BF2E09064032353CFD14B519")
	copy(publicKey[:], publicKeySlice)
}

func main() {
	dhtServer, err := dht.New()
	if err != nil {
		fmt.Printf("Failed to create server %s.\n", err)
		return
	}
	go dhtServer.Serve()
	defer dhtServer.Stop()

	dhtServer.AddFriend(&publicKey)
	dhtServer.AddFriend(&dht.DhtServerList[0].PublicKey)

	dhtServer.Bootstrap(dht.DhtServerList[0])

	data, err := dhtServer.PackPingPong(true, 1, &dht.DhtServerList[0].PublicKey)
	err = dhtServer.Send(data, &dht.DhtServerList[0].Addr)
	if err != nil {
		fmt.Printf("error %s\n", err)
		return
	}

	<-dhtServer.Request
}
