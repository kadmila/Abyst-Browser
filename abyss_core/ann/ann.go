// Package ann (abyss net node) provides QUIC node that can establish
// abyss P2P connections and TLS client auth HTTPS connections.
// This implements ani (abyss new interface) for alpha release.
// TODO: AbyssNodeConfig for construction (backlog, firewall, logger, etc)
package ann

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"time"

	"github.com/kadmila/Abyss-Browser/abyss_core/sec"
	"github.com/quic-go/quic-go"
)

type AbyssNode struct {
	*sec.AbyssRootSecret
	*sec.TLSIdentity

	verified_tls_certs    *sec.VerifiedTlsCertMap
	transport             *quic.Transport
	listener              *quic.Listener
	local_addr_candidates []netip.AddrPort

	inner_ctx  context.Context
	ctx_cancel context.CancelFunc
	inner_done chan bool
	backlog    chan *AbyssPeer
}

func NewAbyssNode(root_private_key sec.PrivateKey) (*AbyssNode, error) {
	root_secret, err := sec.NewAbyssRootSecrets(root_private_key)
	if err != nil {
		return nil, err
	}

	tls_identity, err := root_secret.NewTLSIdentity()
	if err != nil {
		return nil, err
	}

	verified_tls_certs := sec.NewVerifiedTlsCertMap()

	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	transport := &quic.Transport{Conn: udpConn}

	listener, err := transport.Listen(tls_identity.NewServerTlsConf(verified_tls_certs), NewQuicConf())
	if err != nil {
		return nil, err
	}

	bind_addr, ok := listener.Addr().(*net.UDPAddr)
	if !ok {
		return nil, errors.New("failed to get listener bind address")
	}
	port := uint16(bind_addr.Port)

	var addr_candidates []netip.AddrPort

	// query all network interfaces.
	{
		ifaces, err := net.Interfaces()
		if err != nil {
			return nil, err
		}
		for _, iface := range ifaces {
			// Skip disabled interfaces
			if iface.Flags&net.FlagUp == 0 {
				continue
			}

			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				var ip net.IP
				// sugar - go standard library has varying spec over platforms.
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip == nil || ip.To4() == nil {
					continue
				}

				netip_ip, ok := netip.AddrFromSlice(ip.To4())
				if !ok {
					continue
				}
				addr_candidates = append(
					addr_candidates,
					netip.AddrPortFrom(netip_ip, port),
				)
			}
		}
	}

	inner_ctx, ctx_cancel := context.WithCancel(context.Background())

	result := &AbyssNode{
		AbyssRootSecret: root_secret,
		TLSIdentity:     tls_identity,

		verified_tls_certs:    verified_tls_certs,
		transport:             transport,
		listener:              listener,
		local_addr_candidates: addr_candidates,

		inner_ctx:  inner_ctx,
		ctx_cancel: ctx_cancel,
		backlog:    make(chan *AbyssPeer, 32),
	}
	go result.innerLoop()
	return result, nil
}

func NewQuicConf() *quic.Config {
	return &quic.Config{
		MaxIdleTimeout:  time.Second * 20,
		KeepAlivePeriod: time.Second * 5,
		EnableDatagrams: true,
	}
}

func (n *AbyssNode) innerLoop() {
	for {
		connection, err := n.listener.Accept(n.inner_ctx)
		if err != nil {
			if n.inner_ctx.Done() {
				break
			}
		}
	}
	n.inner_done <- true
}

// func T() {
// 	root_key, err := sec.NewRootPrivateKey()
// 	var n ani.IAbyssNode
// 	n, err = NewAbyssNode(root_key)
// }

func (n *AbyssNode) LocalAddrCandidates() []netip.AddrPort { return n.local_addr_candidates }

func (n *AbyssNode) Accept(ctx context.Context) (*AbyssPeer, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case new_peer, ok := <-n.backlog:
		if !ok {
			return nil, errors.New("AbyssNode closed")
		}
		return new_peer, nil
	}
}

func (n *AbyssNode) Dial() {

}

func (n *AbyssNode) Close() {
	n.ctx_cancel()
	n.listener.Close()

	// wait for inner loop termination and close backlog.
	<-n.inner_done
	close(n.backlog)
}
