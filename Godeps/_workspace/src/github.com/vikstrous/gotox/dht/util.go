package dht

import (
	"net"
)

func addrEq(addr1, addr2 net.UDPAddr) bool {
	if !addr1.IP.Equal(addr2.IP) {
		return false
	}
	if addr1.Port != addr2.Port {
		return false
	}
	return true
}
