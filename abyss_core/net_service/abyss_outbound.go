package net_service

import (
	"bytes"
	"crypto/x509"
	"net"

	"github.com/fxamacker/cbor/v2"
	"github.com/quic-go/quic-go"
)

func (h *BetaNetService) PrepareAbyssOutbound(target *ContextedPeer, addresses []*net.UDPAddr) {
	//watchdog.Info("outbound detected")
	var connection quic.Connection
	var ahmp_encoder *cbor.Encoder
	var err error

	defer func() {
		target.mtx.Lock()
		defer target.mtx.Unlock()

		if err != nil {
			if target.err == nil {
				target.err = err
			}
			target.state = PNCS_CLOSED
		} else {
			switch target.state {
			case PNCS_DISCONNECTED:
				target.state = PNCS_OUTBOUND
				target.outbound_conn = connection
				target.addresses = append(target.addresses, addresses...)
				target.ahmp_encoder = ahmp_encoder
			case PNCS_INBOUND:
				target.state = PNCS_CONNECTED
				target.outbound_conn = connection
				target.addresses = append(target.addresses, addresses...)
				target.ahmp_encoder = ahmp_encoder
				h.abyssPeerCH <- target
			case PNCS_OUTBOUND, PNCS_CONNECTED:
				connection.CloseWithError(ABYSS_ALREADY_CONNECTED, ABYSS_ALREADY_CONNECTED_M)
			case PNCS_CLOSED:
				connection.CloseWithError(ABYSS_EARLY_RECONNECTION, ABYSS_EARLY_RECONNECTION_M)
			}
		}
	}()

	address_selected := h.addressSelector.FilterAddressCandidates(addresses)
	connection, err = h.quicTransport.Dial(target.ctx, address_selected[0], h.abyssTlsConf, h.quicConf)
	if err != nil {
		return
	}

	//get self-signed TLS certificate that the peer presented.
	tls_info := connection.ConnectionState().TLS
	client_tls_cert := tls_info.PeerCertificates[0] //*x509.Certificate, validated

	ahmp_stream, err := connection.OpenStreamSync(target.ctx)
	if err != nil {
		return
	}
	ahmp_encoder = cbor.NewEncoder(ahmp_stream)
	ahmp_decoder := cbor.NewDecoder(ahmp_stream)

	//send {local peer_hash, local tls-abyss binding cert} encrypted with remote handshake key.
	var handshake_1_buf bytes.Buffer
	err = cbor.MarshalToBuffer(h.tlsIdentity.abyss_bind_cert, &handshake_1_buf)
	if err != nil {
		return
	}
	handshake_1_payload, err := target.identity.EncryptHandshake(handshake_1_buf.Bytes())
	if err != nil {
		return
	}
	err = ahmp_encoder.Encode(handshake_1_payload)
	if err != nil {
		return
	}

	//receive accepter-side self-authentication
	var handshake_2_payload []byte
	if err := ahmp_decoder.Decode(&handshake_2_payload); err != nil {
		return
	}
	handshake_2_payload_x509, err := x509.ParseCertificate(handshake_2_payload)
	if err := target.identity.VerifyTLSBinding(handshake_2_payload_x509, client_tls_cert); err != nil {
		return
	}

	//return: defer will update the peer.
}
