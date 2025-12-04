package interfaces

import "net"

type IIpFilter interface {
	IsInboundOpen(addr *net.Addr) bool
}
