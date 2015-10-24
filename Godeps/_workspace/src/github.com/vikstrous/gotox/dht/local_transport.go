package dht

import (
	"log"
	"net"
)

type LocalTransport struct {
	ChOut       *chan []byte
	ChIn        *chan []byte
	ChDone      chan struct{}
	Identity    *Identity
	ReceiveFunc ReceiveFunc
}

func NewLocalTransport(id *Identity) (*LocalTransport, error) {
	chIn := make(chan []byte, 100)
	return &LocalTransport{
		ChIn:     &chIn,
		Identity: id,
		ChDone:   make(chan struct{}),
	}, nil
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

func (t *LocalTransport) Listen(ch chan struct{}) {
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
			terminate := t.ReceiveFunc(plainPacket, &net.UDPAddr{})
			if terminate {
				log.Printf("Clean termination.")
				break listenLoop
			}
		case <-t.ChDone:
			break listenLoop
		}
	}
	if ch != nil {
		close(ch)
	}
	return
}

func (t *LocalTransport) Stop() {
	close(t.ChDone)
}

func (t *LocalTransport) RegisterReceiver(receiver ReceiveFunc) {
	t.ReceiveFunc = receiver
}
