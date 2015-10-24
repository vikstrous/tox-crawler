package dht

import (
	"log"
	"net"
	"time"
)

//  Transport/Listen[with identity] -> Application/Receiver.Receive() ...
//  Transport/Sender[with identity].Send()
//  Application whatever -> Transport/Sender[with identity].Send()

// return true to terminate the listener
type ReceiveFunc func(*PlainPacket, *net.UDPAddr) bool

type Transport interface {
	Send(payload Payload, dest *DHTPeer) error
	Listen(chan struct{})
	Stop()
	RegisterReceiver(receiver ReceiveFunc)
}

type UDPTransport struct {
	Server      net.UDPConn
	Identity    *Identity
	ReceiveFunc ReceiveFunc
	ChDone      chan struct{}
}

func NewUDPTransport(id *Identity) (*UDPTransport, error) {
	listener, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}

	return &UDPTransport{
		Server:   *listener,
		Identity: id,
		ChDone:   make(chan struct{}),
	}, nil
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
	close(t.ChDone)
}

func (t *UDPTransport) Listen(ch chan struct{}) {
listenLoop:
	for {
		select {
		case <-t.ChDone:
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
				terminate := t.ReceiveFunc(plainPacket, addr)
				if terminate {
					log.Printf("Clean termination.")
					break listenLoop
				}
			} else {
				log.Printf("Received empty message???")
				continue
			}
		}
	}
	if ch != nil {
		close(ch)
	}
	return
}

func (t *UDPTransport) RegisterReceiver(receiver ReceiveFunc) {
	t.ReceiveFunc = receiver
}
