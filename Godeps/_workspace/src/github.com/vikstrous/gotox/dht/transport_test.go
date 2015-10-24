package dht

import (
	"net"
	"testing"
)

func TestReceive(t *testing.T) {
	id1, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
		return
	}
	transport1, err := NewLocalTransport(id1)
	if err != nil {
		t.Fatal(err)
		return
	}
	id2, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
		return
	}
	transport2, err := NewLocalTransport(id2)
	if err != nil {
		t.Fatal(err)
		return
	}

	// used to wait for the test to finish
	ch := make(chan struct{})

	// register the receiver before we send
	transport1.RegisterReceiver(func(pp *PlainPacket, addr *net.UDPAddr) bool {
		switch payload := pp.Payload.(type) {
		case *PingPong:
			if payload.IsPing != true {
				t.Fatalf("Was not ping: %b", payload.IsPing)
			}
			if payload.RequestID != 3 {
				t.Fatalf("Wrong pingID: %d", payload.RequestID)
			}
		default:
			t.Fatalf("Internal error. Failed to handle payload of parsed packet. %d\n", pp.Payload.Kind())
		}
		close(ch)
		return true
	})

	// pipe the output of transport2 into the input of transport1
	transport2.ChOut = transport1.ChIn
	transport2.Send(&PingPong{IsPing: true, RequestID: 3}, &DHTPeer{PublicKey: id1.PublicKey})

	// process the message
	transport1.Listen(nil)

	<-ch
}
