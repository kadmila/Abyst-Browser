package and

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type IANDEvent any

type EANDSessionRequest struct {
	SessionID uuid.UUID
	ANDPeerSession
}
type EANDSessionReady struct {
	SessionID uuid.UUID
	ANDPeerSession
}
type EANDSessionClose struct {
	SessionID uuid.UUID
	ANDPeerSession
}
type EANDJoinSuccess struct {
	SessionID uuid.UUID
	URL       string
}
type EANDJoinFail struct {
	SessionID uuid.UUID
	Code      int
	Message   string
}
type EANDWorldLeave struct {
}
type EANDConnectRequest struct {
	PeerID                     string
	AddressCandidates          []netip.AddrPort
	RootCertificateDer         []byte
	HandshakeKeyCertificateDer []byte
}
type EANDTimerRequest struct {
	SessionID uuid.UUID
	Duration  time.Duration
}
type EANDObjectAppend struct {
}
type EANDObjectDelete struct {
}
type EANDError struct {
}
type EANDPeerClose struct {
}
