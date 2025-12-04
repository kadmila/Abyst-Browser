package interfaces

import "net/netip"

// IAbyssNetNode is the interface for abyss net node.
// It can establish connection with known peers.
type IAbyssNetNode interface {
	IHostIdentity
	LocalIPs() []*netip.Addr
	Port() uint16

	GetAbyssPeerChannel() chan IANDPeer
	GetAbystInboundChannel()

	// AppendKnownPeer adds peer information to accept connection.
	AppendKnownPeer(root_cert string, handshake_key_cert string) error
	AppendKnownPeerDer(root_cert []byte, handshake_key_cert []byte) error

	ListenAndServe() error
}
