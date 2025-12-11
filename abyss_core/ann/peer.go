package ann

import (
	"context"
	"net/netip"
)

type AbyssPeer struct {
	AuthenticatedConnection
	origin      *PeerConstructor
	internal_id uint64
}

func NewAbyssPeer(connection AuthenticatedConnection, origin *PeerConstructor, internal_id uint64) *AbyssPeer {
	return &AbyssPeer{
		AuthenticatedConnection: connection,
		origin:                  origin,
		internal_id:             internal_id,
	}
}

func (p *AbyssPeer) RemoteAddr() netip.AddrPort {
	return p.remote_addr
}

func (p *AbyssPeer) Send(v any) error {
	return p.ahmp_encoder.Encode(v)
}
func (p *AbyssPeer) Recv(v any) error {
	return p.ahmp_decoder.Decode(v)
}
func (p *AbyssPeer) Context() context.Context {
	return p.connection.Context()
}

func (p *AbyssPeer) Close() error {
	err := p.connection.CloseWithError(AbyssQuicClose, "")
	p.origin.ReportPeerClose(p)
	return err
}

func (p *AbyssPeer) Equal(subject *AbyssPeer) bool {
	return p.internal_id == subject.internal_id
}
