package and

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
)

type PeerWithLocation struct {
	Peer              ani.IAbyssPeer
	AddressCandidates []netip.AddrPort // this is shared with ANDFullPeerSession
}

type PeerWorldSession struct {
	PeerWithLocation
	SessionID uuid.UUID
	TimeStamp time.Time
}

// ANDFullPeerSessionInfo provides all the information required to
// connect a peer, identify its world session, negotiate ordering.
// As a result, a peer who receives this can construct ANDFullPeerSession.
type ANDFullPeerSessionInfo struct {
	PeerID                     string
	AddressCandidates          []netip.AddrPort
	SessionID                  uuid.UUID
	TimeStamp                  time.Time
	RootCertificateDer         []byte
	HandshakeKeyCertificateDer []byte
}

// ANDPeerSessionIdentity is used to indirect a peer world session,
// which is used in connection recovery.
type ANDPeerSessionIdentity struct {
	PeerID    string
	SessionID uuid.UUID
}

// ANDPeerSession is used to indirect a peer world session
// in IANDEvent.
type ANDPeerSession struct {
	Peer      ani.IAbyssPeer
	SessionID uuid.UUID
}

type ObjectInfo struct {
	ID        uuid.UUID
	Addr      string
	Transform [7]float32
}
