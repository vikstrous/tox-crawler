package dht

import (
	"github.com/vikstrous/gotox"
)

type DHTID [gotox.PublicKeySize]byte

const fakeFriendNumber = 4

const maxUDPPacketSize = 2048

// http://lxr.free-electrons.com/source/include/linux/socket.h
// TODO: get these from the source
const AF_INET byte = 2
const AF_INET6 byte = 10
const TCP_INET byte = 130
const TCP_INET6 byte = 138

// message type, ip, port, public key
const packedNodeSizeIPv4 = 1 + 4 + 2 + gotox.PublicKeySize
const packedNodeSizeIPv6 = 1 + 16 + 2 + gotox.PublicKeySize

var packedNodeSize = map[byte]uint{
	AF_INET:   packedNodeSizeIPv4,
	AF_INET6:  packedNodeSizeIPv6,
	TCP_INET:  packedNodeSizeIPv4,
	TCP_INET6: packedNodeSizeIPv6,
}

const netPacketPing uint8 = 0 /* Ping request packet ID. */
const netPacketPong uint8 = 1 /* Ping response packet ID. */
// Node == DHTPeer
const netPacketGetNodes uint8 = 2 /* Get nodes request packet ID. */
// AKA SendNodesIPv6
const netPacketGetNodesReply uint8 = 4   /* Send nodes response packet ID for other addresses. */
const netPacketCookieRequest uint8 = 24  /* Cookie request packet */
const netPacketCookieResponse uint8 = 25 /* Cookie response packet */
const netPacketCryptoHs uint8 = 26       /* Crypto handshake packet */
const netPacketCryptoData uint8 = 27     /* Crypto data packet */
const netPacketCrypto uint8 = 32         /* Encrypted data packet ID. */
const netPacketLanDiscovery uint8 = 33   /* LAN discovery packet ID. */

/* See:  docs/Prevent_Tracking.txt and onion.{c, h} */
const netPacketOnionSendInitial uint8 = 128
const netPacketOnionSend1 uint8 = 129
const netPacketOnionSend2 uint8 = 130

const netPacketAnnounceRequest uint8 = 131
const netPacketAnnounceResponse uint8 = 132
const netPacketOnionDataRequest uint8 = 133
const netPacketOnionDataResponse uint8 = 134

const netPacketOnionRecv3 uint8 = 140
const netPacketOnionRecv2 uint8 = 141
const netPacketOnionRecv1 uint8 = 142

/* Only used for bootstrap nodes */
const bootstrapInfoPacketId = 240

const toxPortrangeFrom = 33445
const toxPortrangeTo = 33545
const toxPortDefault = toxPortrangeFrom

/* TCP related */
//const tcpOnionFamily = (afInet6 + 1)
//const tcpInet = (afInet6 + 2)
//const tcpInet6 = (afInet6 + 3)
//const tcpFamily = (afInet6 + 4)
