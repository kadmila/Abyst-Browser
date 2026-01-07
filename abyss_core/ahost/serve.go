package ahost

import (
	"errors"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"github.com/kadmila/Abyss-Browser/abyss_core/ahmp"
	"github.com/kadmila/Abyss-Browser/abyss_core/and"
	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
)

type parsibleAhmp[T any] interface {
	TryParse() (*T, error)
}

func tryParseAhmp[RawT parsibleAhmp[T], T any](msg *ahmp.AHMPMessage) (*T, error) {
	var raw RawT
	if err := cbor.Unmarshal(msg.Payload, &raw); err != nil {
		return nil, err
	}
	return raw.TryParse()
}

func (h *AbyssHost) servePeer(peer ani.IAbyssPeer) error {
	participating_worlds := make(map[uuid.UUID]*and.World)
	// We hold this reference to skip h.peer_participating_worlds loopkup.
	// Still, it should be noted that accessing this local variable is not thread-safe.
	h.mtx.Lock()
	h.peer_participating_worlds[peer.ID()] = participating_worlds
	h.mtx.Unlock()
	events := and.NewANDEventQueue()

	defer func() {
		h.mtx.Lock()
		defer h.mtx.Unlock()

		for _, world := range participating_worlds {
			world.PeerDisconnected(events, peer.ID())
			h.handleANDEvent(events)
		}
		peer.Close()
	}()

	var msg ahmp.AHMPMessage
	for {
		err := peer.Recv(&msg)
		if err != nil {
			return err
		}
		//fmt.Println("recv: ", msg.Type.String(), h.ID())
		switch msg.Type {
		case ahmp.JN_T:
			JN, err := tryParseAhmp[*and.RawJN](&msg)
			if err != nil {
				return err
			}
			if err := h.onJN(events, JN, and.ANDPeerSession{Peer: peer, SessionID: JN.SenderSessionID}, participating_worlds); err != nil {
				return err
			}
		case ahmp.JOK_T:
			JOK, err := tryParseAhmp[*and.RawJOK](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: JOK.SenderSessionID}
			if err := h.onJOK(events, JOK, peer_session, participating_worlds); err != nil {
				return err
			}
		case ahmp.JDN_T:
			JDN, err := tryParseAhmp[*and.RawJDN](&msg)
			if err != nil {
				return err
			}
			if err := h.onJDN(events, JDN, peer, participating_worlds); err != nil {
				return err
			}
		case ahmp.JNI_T:
			JNI, err := tryParseAhmp[*and.RawJNI](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: JNI.SenderSessionID}
			if err := h.onJNI(events, JNI, peer_session, participating_worlds, JNI.Neighbor); err != nil {
				return err
			}
		case ahmp.MEM_T:
			MEM, err := tryParseAhmp[*and.RawMEM](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: MEM.SenderSessionID}
			if err := h.onMEM(events, MEM, peer_session, participating_worlds); err != nil {
				return err
			}
		case ahmp.SJN_T:
			SJN, err := tryParseAhmp[*and.RawSJN](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: SJN.SenderSessionID}
			if err := h.onSJN(events, SJN, peer_session, participating_worlds); err != nil {
				return err
			}
		case ahmp.CRR_T:
			CRR, err := tryParseAhmp[*and.RawCRR](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: CRR.SenderSessionID}
			if err := h.onCRR(events, CRR, peer_session, participating_worlds); err != nil {
				return err
			}
		case ahmp.RST_T:
			RST, err := tryParseAhmp[*and.RawRST](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: RST.SenderSessionID}
			if err := h.onRST(events, RST, peer_session, participating_worlds); err != nil {
				return err
			}
		case ahmp.SOA_T:
			SOA, err := tryParseAhmp[*and.RawSOA](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: SOA.SenderSessionID}
			if err := h.onSOA(events, SOA, peer_session, participating_worlds); err != nil {
				return err
			}
		case ahmp.SOD_T:
			SOD, err := tryParseAhmp[*and.RawSOD](&msg)
			if err != nil {
				return err
			}
			peer_session := and.ANDPeerSession{Peer: peer, SessionID: SOD.SenderSessionID}
			if err := h.onSOD(events, SOD, peer_session, participating_worlds); err != nil {
				return err
			}
		case ahmp.AU_PING_TX_T:
			if err := h.onAUPingTX(events, peer); err != nil {
				return err
			}
		case ahmp.AU_PING_RX_T:
			if err := h.onAUPingRX(events, peer); err != nil {
				return err
			}
		default:
			// malformed message
			return errors.New("unsupported AHMP message type")
		}
	}
}
