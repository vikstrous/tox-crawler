package dht

import (
	"log"
	"net"
)

type LocalTransport struct {
	ChOut       *chan []byte
	ChIn        *chan []byte
	ChStop      chan struct{}
	Identity    *Identity
	ReceiveFunc ReceiveFunc
	dataChan    chan TransportMessage
}

func NewLocalTransport(id *Identity) (*LocalTransport, error) {
	chIn := make(chan []byte, 100)
	return &LocalTransport{
		ChIn:     &chIn,
		Identity: id,
		ChStop:   make(chan struct{}),
		dataChan: make(chan TransportMessage, 100),
	}, nil
}

func (t *LocalTransport) DataChan() chan TransportMessage {
	return t.dataChan
}

func (t *LocalTransport) Send(payload Payload, dest *DHTPeer) error {
	plainPacket := PlainPacket{
		Sender:  &t.Identity.PublicKey,
		Payload: payload,
	}

	encryptedPacket, err := t.Identity.EncryptPacket(&plainPacket, &dest.PublicKey)
	if err != nil {
		return err
	}

	data, err := encryptedPacket.MarshalBinary()
	if err != nil {
		return err
	}

	*t.ChOut <- data

	return nil
}

func (t *LocalTransport) Listen() {
listenLoop:
	for {
		select {
		case data := <-*t.ChIn:
			var encryptedPacket EncryptedPacket
			err := encryptedPacket.UnmarshalBinary(data)
			if err != nil {
				log.Printf("error receiving: %v", err)
				continue
			}
			plainPacket, err := t.Identity.DecryptPacket(&encryptedPacket)
			if err != nil {
				log.Printf("error receiving: %v", err)
				continue
			}
			t.dataChan <- TransportMessage{*plainPacket, net.UDPAddr{}}
		case <-t.ChStop:
			break listenLoop
		}
	}
	close(t.dataChan)
	return
}

func (t *LocalTransport) Stop() {
	close(t.ChStop)
}

func (t *LocalTransport) RegisterReceiver(receiver ReceiveFunc) {
	t.ReceiveFunc = receiver
}
