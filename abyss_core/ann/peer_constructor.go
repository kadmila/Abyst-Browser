package ann

import (
	"context"
	"net/netip"
)

// PeerConstructor handles appended peer/error, and writes to BackLog.
// When BackLog is full, Append() and AppendError() will block.
type PeerConstructor struct {
	BackLog chan *AbyssPeer

	authenticated_backlog chan *AuthenticatedConnection
}

func NewPeerConstructor() *PeerConstructor {
	return &PeerConstructor{
		authenticated_backlog: make(chan *AuthenticatedConnection, 128),
	}
}

// Append blocks until 1) context cancels, or 2) abyss peer is constructed.
// * Issue: it hard-blocks when BackLog is full.
func (c *PeerConstructor) Append(ctx context.Context, connection *AuthenticatedConnection) {

}

func (c *PeerConstructor) AppendError(addr netip.AddrPort, is_dialing bool, err error) {

}

// consumePendingInboundOrRegisterOutbound returns (completed_connection, ok, inbound_wait, redundant)
// func (n *AbyssNode) consumePendingInboundOrRegisterOutbound(id string, connection quic.Connection) (*AbyssConnection, bool, chan *InboundConnection, bool) {
// 	n.backlog_join_mtx.Lock()
// 	defer n.backlog_join_mtx.Unlock()

// 	if _, ok := n.peers[id]; ok {
// 		return nil, false, nil, true
// 	}
// 	if _, ok := n.outbound_backlog[id]; ok {
// 		return nil, false, nil, true
// 	}

// 	inbound, ok := n.inbound_backlog[id]
// 	if ok {
// 		delete(n.inbound_backlog, id)
// 		return &AbyssConnection{
// 			inbound_connection: inbound.conn,
// 			outbound_connection: connection,
// 			ahmp_encoder: ,
// 		}, true, nil, false
// 	}

// 	inbound_wait := make(chan *InboundConnection)
// 	n.outbound_backlog[id] = inbound_wait
// 	return nil, false, inbound_wait, false
// }

// func (n *AbyssNode) isDialRedundant(id string) bool {
// 	n.backlog_join_mtx.Lock()
// 	defer n.backlog_join_mtx.Unlock()

// 	_, ok := n.outbound_backlog[id]
// 	if ok {
// 		return true
// 	}

// 	_, ok = n.peers[id]
// 	if ok {
// 		return true
// 	}

// 	return false
// }

// func (n *AbyssNode) OutboundConnectionJoin(id string, connection quic.Connection, identity *sec.AbyssPeerIdentity) {
// 	n.backlog_join_mtx.Lock()
// 	defer n.backlog_join_mtx.Unlock()

// 	if _, ok := n.peers[id]; ok {
// 		conn.CloseWithError(AbyssQuicRedundantConnection, "redundant connection")
// 		return
// 	}

// 	outbound_conn, ok := n.outbound_backlog[id]
// 	if ok {
// 		conn.CloseWithError(AbyssQuicRedundantConnection, "redundant connection")
// 	}

// 	inbound_conn, ok := n.inbound_backlog[id]
// 	if ok {

// 	} else {
// 		n.outbound_backlog[id]
// 	}
// }

// TODO func (n *AbyssNode) NewAbystClient() (IAbystClient, error) {}

// TODO NewCollocatedHttp3Client() (http.Client, error)
