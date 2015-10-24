package dht

import (
	//"encoding/hex"
	"net"
	"reflect"
	"testing"

	//"github.com/vikstrous/gotox"
	//"golang.org/x/crypto/nacl/box"
)

var TestIdentity *Identity
var Addr4 net.UDPAddr
var Addr6 net.UDPAddr

func init() {
	TestIdentity, _ = GenerateIdentity()
	Addr6.IP = net.ParseIP("::1")
	Addr6.Port = 1337
	Addr4.IP = net.ParseIP("127.0.0.1")
	Addr4.Port = 1337
}

func testEncryptDecrypt(t *testing.T, id *Identity, pp *PlainPacket) {
	ep, err := id.EncryptPacket(pp, &id.PublicKey)
	if err != nil {
		t.Fatalf("Failed to encrypt. %s", err)
	}
	data, err := ep.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal encrypted data. %s", err)
	}
	ep2 := &EncryptedPacket{}
	err = ep2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal encrypted data. %s", err)
	}
	pp2, err := id.DecryptPacket(ep2)
	if err != nil {
		t.Fatalf("Failed to decrypt encrypted packet. %s", err)
	}
	if !reflect.DeepEqual(pp2, pp) {
		t.Fatalf("Mismatch after decryption.\n %v\n from %v\n VS\n %v\n from %v\n", pp.Payload, pp.Sender, pp2.Payload, pp2.Sender)
	}
}

func TestPing(t *testing.T) {
	pp := &PlainPacket{Sender: &TestIdentity.PublicKey, Payload: &PingPong{true, 1}}
	testEncryptDecrypt(t, TestIdentity, pp)
}

func TestPong(t *testing.T) {
	pp := &PlainPacket{Sender: &TestIdentity.PublicKey, Payload: &PingPong{false, 2}}
	testEncryptDecrypt(t, TestIdentity, pp)
}

func TestGetNodes(t *testing.T) {
	pp := &PlainPacket{Sender: &TestIdentity.PublicKey, Payload: &GetNodes{&TestIdentity.PublicKey, 3}}
	testEncryptDecrypt(t, TestIdentity, pp)

}

func TestGetNodesReply1(t *testing.T) {
	pp := &PlainPacket{Sender: &TestIdentity.PublicKey, Payload: &GetNodesReply{[]DHTPeer{{TestIdentity.PublicKey, Addr4}}, 4}}
	testEncryptDecrypt(t, TestIdentity, pp)
}

func TestGetNodesReply2(t *testing.T) {
	pp := &PlainPacket{Sender: &TestIdentity.PublicKey, Payload: &GetNodesReply{[]DHTPeer{{TestIdentity.PublicKey, Addr4}, {TestIdentity.PublicKey, Addr6}}, 5}}
	testEncryptDecrypt(t, TestIdentity, pp)
}
func TestGetNodesReply5(t *testing.T) {
	node := DHTPeer{TestIdentity.PublicKey, Addr4}
	pp := &PlainPacket{Sender: &TestIdentity.PublicKey, Payload: &GetNodesReply{[]DHTPeer{
		node,
		node,
		node,
		node,
		node},
		5}}
	_, err := TestIdentity.EncryptPacket(pp, &TestIdentity.PublicKey)
	if err == nil {
		t.Fatalf("Should not have succeed in encrypting with too many nodes in GetNodes.")
	}
}
