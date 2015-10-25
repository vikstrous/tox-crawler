package dht

import (
	"log"
	"net"
	"time"
)

// return true to terminate the listener
type ReceiveFunc func(*PlainPacket, *net.UDPAddr) bool

type TransportMessage struct {
	Packet PlainPacket
	Addr   net.UDPAddr
}

type Transport interface {
	Send(payload Payload, dest *DHTPeer) error
	Listen()
	Stop()
	DataChan() chan TransportMessage
}

type UDPTransport struct {
	Server   net.UDPConn
	Identity *Identity
	dataChan chan TransportMessage
	ChStop   chan struct{}
}

func NewUDPTransport(id *Identity) (*UDPTransport, error) {
	listener, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}

	return &UDPTransport{
		Server:   *listener,
		Identity: id,
		ChStop:   make(chan struct{}),
		dataChan: make(chan TransportMessage, 100),
	}, nil
}

func (t *UDPTransport) DataChan() chan TransportMessage {
	return t.dataChan
}

func (t *UDPTransport) Send(payload Payload, dest *DHTPeer) error {
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

	t.Server.SetDeadline(time.Now().Add(time.Second))
	_, _, err = t.Server.WriteMsgUDP(data, nil, &dest.Addr)
	return err
}

func (t *UDPTransport) Stop() {
	close(t.ChStop)
}

func (t *UDPTransport) Listen() {
listenLoop:
	for {
		select {
		case <-t.ChStop:
			break listenLoop
		default:
			buffer := make([]byte, 2048)
			// TODO: can we make this buffer smaller?
			oob := make([]byte, 2048)
			// n, oobn, flags, addr, err
			t.Server.SetDeadline(time.Now().Add(time.Second))
			buffer_length, _, _, addr, err := t.Server.ReadMsgUDP(buffer, oob)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				log.Println(err)
				break listenLoop
			}

			if buffer_length >= 1 {
				var encryptedPacket EncryptedPacket
				err := encryptedPacket.UnmarshalBinary(buffer[:buffer_length])
				if err != nil {
					log.Printf("error receiving: %v", err)
					continue
				}
				plainPacket, err := t.Identity.DecryptPacket(&encryptedPacket)
				if err != nil {
					log.Printf("error receiving: %v", err)
					continue
				}
				t.dataChan <- TransportMessage{*plainPacket, *addr}
			} else {
				log.Printf("Received empty message???")
				continue
			}
		}
	}
	close(t.dataChan)
	return
}
