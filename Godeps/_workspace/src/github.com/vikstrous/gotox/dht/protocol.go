package dht

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"

	"github.com/vikstrous/gotox"
)

var DhtServerList []DHTPeer

type DHTPeer struct {
	PublicKey [gotox.PublicKeySize]byte
	Addr      net.UDPAddr
}

func init() {
	serverListJSON := `[
	{
		"Name":"sonfOfRa",
		"PublicKey":"04119E835DF3E78BACF0F84235B300546AF8B936F035185E2A8E9E0A67C8924F",
		"Addr":{"IP":"144.76.60.215","Port":33445}
	},
	{
		"Name":"sta1",
		"PublicKey":"A09162D68618E742FFBCA1C2C70385E6679604B2D80EA6E84AD0996A1AC8A074",
		"Addr":{"IP":"23.226.230.47","Port":33445}
	},
	{
		"Name":"Munrek",
		"PublicKey":"E398A69646B8CEACA9F0B84F553726C1C49270558C57DF5F3C368F05A7D71354",
		"Addr":{"IP":"195.154.119.113","Port":33445}
	},
	{
		"Name":"nurupo",
		"PublicKey":"F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67",
		"Addr":{"IP":"192.210.149.121","Port":33445}
	},
	{
		"Name":"Impyy",
		"PublicKey":"788236D34978D1D5BD822F0A5BEBD2C53C64CC31CD3149350EE27D4D9A2F9B6B",
		"Addr":{"IP":"178.62.250.138","Port":33445}
	},
	{
		"Name":"Manolis",
		"PublicKey":"461FA3776EF0FA655F1A05477DF1B3B614F7D6B124F7DB1DD4FE3C08B03B640F",
		"Addr":{"IP":"130.133.110.14","Port":33445}
	},
	{
		"Name":"noisykeyboard",
		"PublicKey":"5918AC3C06955962A75AD7DF4F80A5D7C34F7DB9E1498D2E0495DE35B3FE8A57",
		"Addr":{"IP":"104.167.101.29","Port":33445}
	},
	{
		"Name":"Busindre",
		"PublicKey":"A179B09749AC826FF01F37A9613F6B57118AE014D4196A0E1105A98F93A54702",
		"Addr":{"IP":"205.185.116.116","Port":33445}
	},
	{
		"Name":"Busindre",
		"PublicKey":"1D5A5F2F5D6233058BF0259B09622FB40B482E4FA0931EB8FD3AB8E7BF7DAF6F",
		"Addr":{"IP":"198.98.51.198","Port":33445}
	},
	{
		"Name":"ray65536",
		"PublicKey":"8E7D0B859922EF569298B4D261A8CCB5FEA14FB91ED412A7603A585A25698832",
		"Addr":{"IP":"108.61.165.198","Port":33445}
	},
	{
		"Name":"Kr9r0x",
		"PublicKey":"C4CEB8C7AC607C6B374E2E782B3C00EA3A63B80D4910B8649CCACDD19F260819",
		"Addr":{"IP":"212.71.252.109","Port":33445}
	},
	{
		"Name":"fluke571",
		"PublicKey":"3CEE1F054081E7A011234883BC4FC39F661A55B73637A5AC293DDF1251D9432B",
		"Addr":{"IP":"194.249.212.109","Port":33445}
	},
	{
		"Name":"MAH69K",
		"PublicKey":"DA4E4ED4B697F2E9B000EEFE3A34B554ACD3F45F5C96EAEA2516DD7FF9AF7B43",
		"Addr":{"IP":"185.25.116.107","Port":33445}
	},
	{
		"Name":"WIeschie",
		"PublicKey":"6A4D0607A296838434A6A7DDF99F50EF9D60A2C510BBF31FE538A25CB6B4652F",
		"Addr":{"IP":"192.99.168.140","Port":33445}
	}
]`

	json.Unmarshal([]byte(serverListJSON), &DhtServerList)
}

func (n *DHTPeer) UnmarshalJSON(data []byte) error {
	tmp := struct {
		Name      string
		PublicKey string
		Addr      net.UDPAddr
	}{}
	err := json.Unmarshal([]byte(data), &tmp)
	if err != nil {
		return err
	}
	publicKey, err := hex.DecodeString(tmp.PublicKey)
	if err != nil {
		return err
	}
	copy(n.PublicKey[:], publicKey)
	n.Addr = tmp.Addr
	return nil
}

type GetNodes struct {
	RequestedNodeID *[gotox.PublicKeySize]byte
	RequestID       uint64
}

type GetNodesReply struct {
	Nodes        []DHTPeer
	SendbackData uint64
}

type EncryptedPacket struct {
	Kind    uint8
	Sender  *[gotox.PublicKeySize]byte
	Nonce   *[gotox.NonceSize]byte
	Payload []byte
}

type Payload interface {
	Kind() uint8
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}

type PlainPacket struct {
	Sender  *[gotox.PublicKeySize]byte
	Payload Payload
}

type PingPong struct {
	IsPing    bool
	RequestID uint64
}

// packedNodeSizeIp6
// 1 + 32 + 24 + 1 + n*(32+1+16+2) + 8 + overhead
//
// [byte with value: 04]
// [char array  (client node_id), length=32 bytes]
// [random 24 byte nonce]
// [Encrypted with the nonce and private key of the sender:
//     [uint8_t number of nodes in this packet]
//     [Nodes in node format, length=?? * (number of nodes (maximum of 4 nodes)) bytes]
//     [Sendback data, length=8 bytes]
// ]

//[char array (node_id), length=32 bytes]
//[uint8_t family (2 == IPv4, 10 == IPv6, 130 == TCP IPv4, 138 == TCP IPv6)]
//[ip (in network byte order), length=4 bytes if ipv4, 16 bytes if ipv6]
//[port (in network byte order), length=2 bytes]

func (sn *GetNodesReply) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	if len(sn.Nodes) > 4 {
		return nil, fmt.Errorf("Attempt to send too many nodes in reply: %d", len(sn.Nodes))
	}

	// number
	err := binary.Write(buf, binary.BigEndian, uint8(len(sn.Nodes)))
	if err != nil {
		return nil, err
	}

	// nodes
	for _, node := range sn.Nodes {
		nodeBytes, err := node.MarshalBinary()
		if err != nil {
			return nil, err
		}

		_, err = buf.Write(nodeBytes)
		if err != nil {
			return nil, err
		}
	}

	// sendback data
	err = binary.Write(buf, binary.BigEndian, sn.SendbackData)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (sn *GetNodesReply) Kind() uint8 {
	return netPacketGetNodesReply
}

func (sn *GetNodesReply) UnmarshalBinary(data []byte) error {
	//log.Printf("GetNodesReply data %v %d", data, len(data))
	// number of nodes
	numNodes := uint8(len(sn.Nodes))
	binary.Read(bytes.NewReader(data), binary.BigEndian, &numNodes)

	// nodes
	sn.Nodes = make([]DHTPeer, numNodes)
	offset := 1
	for n := uint8(0); n < numNodes; n++ {
		var nodeSize int
		if data[offset] == AF_INET || data[offset] == TCP_INET {
			nodeSize = packedNodeSizeIPv4
		} else if data[offset] == AF_INET6 || data[offset] == TCP_INET6 {
			nodeSize = packedNodeSizeIPv6
		} else {
			return fmt.Errorf("Unknown ip type %d", data[offset])
		}
		err := sn.Nodes[n].UnmarshalBinary(data[offset : offset+nodeSize])
		if err != nil {
			return err
		}
		offset += nodeSize
	}

	if len(data) != offset+8 {
		return fmt.Errorf("Wrong length packet decrypted! Expected %d, got %d.", offset+8, len(data))
	}

	// sendback data
	return binary.Read(bytes.NewReader(data[offset:]), binary.BigEndian, &sn.SendbackData)
}

func (n *DHTPeer) MarshalBinary() ([]byte, error) {
	// TODO: support TCP
	buf := new(bytes.Buffer)
	var err error
	if ipv4 := n.Addr.IP.To4(); ipv4 != nil {
		// family 1 byte
		err = binary.Write(buf, binary.BigEndian, AF_INET)
		if err != nil {
			return nil, err
		}
		// address 4 bytes
		_, err = buf.Write(ipv4)
		if err != nil {
			return nil, err
		}
	} else if ipv6 := n.Addr.IP.To16(); ipv6 != nil {
		// family 1 byte
		err = binary.Write(buf, binary.BigEndian, AF_INET6)
		if err != nil {
			return nil, err
		}
		// address 16 bytes
		_, err = buf.Write(ipv6)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Invalid node address for node %v", n)
	}
	// port 2 bytes
	err = binary.Write(buf, binary.BigEndian, uint16(n.Addr.Port))
	if err != nil {
		return nil, err
	}
	// public key 32 bytes
	_, err = buf.Write(n.PublicKey[:])
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (n *DHTPeer) UnmarshalBinary(data []byte) error {
	if len(data) != packedNodeSizeIPv4 && len(data) != packedNodeSizeIPv6 {
		return fmt.Errorf("Wrong size data for node %d", len(data))
	}

	//log.Printf("parsing %v %d", data, len(data))
	// ip type
	ipType := data[0]

	// confirm ip type
	var ipSize int
	if ipType == AF_INET || ipType == TCP_INET {
		ipSize = net.IPv4len
		n.Addr.IP = net.IPv4(data[1], data[2], data[3], data[4])
	} else if ipType == AF_INET6 || ipType == TCP_INET6 {
		ipSize = net.IPv6len
		n.Addr.IP = data[1 : ipSize+1]
	} else {
		return fmt.Errorf("Unknown ip type %d", ipType)
	}

	// port
	var port uint16
	err := binary.Read(bytes.NewReader(data[1+ipSize:]), binary.BigEndian, &port)
	if err != nil {
		return err
	}
	n.Addr.Port = int(port)

	// public key
	copy(n.PublicKey[:], data[1+ipSize+2:])
	return nil
}

func (p *EncryptedPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 1+gotox.PublicKeySize+gotox.NonceSize {
		return fmt.Errorf("Packet too small to be valid %d", len(data))
	}
	p.Kind = uint8(data[0])
	p.Sender = new([gotox.PublicKeySize]byte)
	copy(p.Sender[:], data[1:1+gotox.PublicKeySize])
	p.Nonce = new([gotox.NonceSize]byte)
	copy(p.Nonce[:], data[1+gotox.PublicKeySize:1+gotox.PublicKeySize+gotox.NonceSize])
	p.Payload = data[1+gotox.PublicKeySize+gotox.NonceSize:]
	return nil
}

func (p *EncryptedPacket) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 1 byte message type
	err := binary.Write(buf, binary.BigEndian, p.Kind)
	if err != nil {
		return nil, err
	}

	// 32 byte public key
	_, err = buf.Write(p.Sender[:])
	if err != nil {
		return nil, err
	}

	// write the nonce
	_, err = buf.Write(p.Nonce[:])
	if err != nil {
		return nil, err
	}

	// write the encrypted message
	_, err = buf.Write(p.Payload)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *PingPong) MarshalBinary() ([]byte, error) {
	data := new(bytes.Buffer)
	var kind uint8
	if p.IsPing {
		kind = netPacketPing
	} else {
		kind = netPacketPong
	}
	// request or respense
	err := binary.Write(data, binary.BigEndian, kind)
	if err != nil {
		return nil, err
	}
	// pind id
	err = binary.Write(data, binary.BigEndian, p.RequestID)
	if err != nil {
		return nil, err
	}
	// finalize message to be encrypted
	return data.Bytes(), nil
}

func (p *PingPong) Kind() uint8 {
	if p.IsPing {
		return netPacketPing
	} else {
		return netPacketPong
	}
}

func (p *PingPong) UnmarshalBinary(data []byte) error {
	if len(data) < 1+8 {
		return fmt.Errorf("Wrong size data for ping %d.", len(data))
	}
	if data[0] == netPacketPing {
		p.IsPing = true
	} else if data[0] == netPacketPong {
		p.IsPing = false
	} else {
		return fmt.Errorf("Unknown ping type %d.", data[0])
	}
	return binary.Read(bytes.NewReader(data[1:]), binary.BigEndian, &p.RequestID)
}

func (sn *GetNodes) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	// node id
	_, err := buf.Write(sn.RequestedNodeID[:])
	if err != nil {
		return nil, err
	}
	// sendback data
	err = binary.Write(buf, binary.BigEndian, sn.RequestID)
	if err != nil {
		return nil, err
	}
	// finalize message to be encrypted
	return buf.Bytes(), nil
}

func (sn *GetNodes) Kind() uint8 {
	return netPacketGetNodes
}

func (sn *GetNodes) UnmarshalBinary(data []byte) error {
	//TODO: check length
	sn.RequestedNodeID = new([gotox.PublicKeySize]byte)
	copy(sn.RequestedNodeID[:], data[:gotox.PublicKeySize])
	return binary.Read(bytes.NewReader(data[gotox.PublicKeySize:gotox.PublicKeySize+gotox.SendbackDataSize]), binary.BigEndian, &sn.RequestID)
}
