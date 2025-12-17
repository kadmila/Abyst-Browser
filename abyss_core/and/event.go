package and

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type IANDEvent any

type EANDSessionRequest struct {
	world *World
	ANDPeerSession
}
type EANDSessionReady struct {
	world *World
	ANDPeerSession
}
type EANDSessionClose struct {
	world *World
	ANDPeerSession
}
type EANDJoinSuccess struct {
	world *World
	URL   string
}
type EANDJoinFail struct {
	world   *World
	Code    int
	Message string
}
type EANDWorldLeave struct {
	world *World
}
type EANDConnectRequest struct {
	PeerID                     string
	AddressCandidates          []netip.AddrPort
	RootCertificateDer         []byte
	HandshakeKeyCertificateDer []byte
}
type EANDTimerRequest struct {
	world    *World
	Duration time.Duration
}
type EANDObjectAppend struct {
	world *World
	ANDPeerSession
	Objects []ObjectInfo
}
type EANDObjectDelete struct {
	world *World
	ANDPeerSession
	ObjectIDs []uuid.UUID
}
type EANDError struct {
}
