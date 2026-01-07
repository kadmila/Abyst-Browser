package ahost

import (
	"github.com/google/uuid"
	"github.com/kadmila/Abyss-Browser/abyss_core/and"
	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
)

func (h *AbyssHost) onJN(
	events *and.ANDEventQueue,
	JN *and.JN,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := h.exposed_worlds[JN.Path]
	if !ok {
		return and.SendJDN_NoWorld(peer_session, and.JNC_NOT_FOUND, and.JNM_NOT_FOUND)
	}

	// JN forces appending participating_worlds.
	if _, ok := participating_worlds[world.SessionID()]; !ok {
		participating_worlds[world.SessionID()] = world
		world.PeerConnected(events, peer_session.Peer)
		h.handleANDEvent(events)
	}

	world.JN(events, peer_session, JN.TimeStamp)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onJOK(
	events *and.ANDEventQueue,
	JOK *and.JOK,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := participating_worlds[JOK.RecverSessionID]
	if !ok {
		return nil
	}
	world.JOK(events, peer_session, JOK.TimeStamp, JOK.URL, JOK.Neighbors)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onJDN(
	events *and.ANDEventQueue,
	JDN *and.JDN,
	peer ani.IAbyssPeer,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := participating_worlds[JDN.RecverSessionID]
	if !ok {
		return nil
	}
	world.JDN(events, peer, JDN.Code, JDN.Message)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onJNI(
	events *and.ANDEventQueue,
	JNI *and.JNI,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
	joiner_info and.ANDFullPeerSessionInfo,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	world, ok := participating_worlds[JNI.RecverSessionID]
	if !ok {
		return nil
	}
	world.JNI(events, peer_session, joiner_info)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onMEM(
	events *and.ANDEventQueue,
	MEM *and.MEM,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	// MEM forces appending participating_worlds.
	world, ok := participating_worlds[MEM.RecverSessionID]
	if !ok {
		world, ok = h.worlds[MEM.RecverSessionID]
		if !ok {
			return nil
		}

		participating_worlds[world.SessionID()] = world
		world.PeerConnected(events, peer_session.Peer)
		h.handleANDEvent(events)
	}

	world.MEM(events, peer_session, MEM.TimeStamp)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onSJN(
	events *and.ANDEventQueue,
	SJN *and.SJN,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := participating_worlds[SJN.RecverSessionID]
	if !ok {
		return nil
	}
	world.SJN(events, peer_session, SJN.MemberInfos)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onCRR(
	events *and.ANDEventQueue,
	CRR *and.CRR,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := participating_worlds[CRR.RecverSessionID]
	if !ok {
		return nil
	}
	world.CRR(events, peer_session, CRR.MemberInfos)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onRST(
	events *and.ANDEventQueue,
	RST *and.RST,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := participating_worlds[RST.RecverSessionID]
	if !ok {
		return nil // resetting non-participating world is a no-op.
	}
	world.RST(events, peer_session)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onSOA(
	events *and.ANDEventQueue,
	SOA *and.SOA,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := participating_worlds[SOA.RecverSessionID]
	if !ok {
		return nil
	}
	world.SOA(events, peer_session, SOA.Objects)
	h.handleANDEvent(events)
	return nil
}

func (h *AbyssHost) onSOD(
	events *and.ANDEventQueue,
	SOD *and.SOD,
	peer_session and.ANDPeerSession,
	participating_worlds map[uuid.UUID]*and.World,
) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	world, ok := participating_worlds[SOD.RecverSessionID]
	if !ok {
		return nil
	}
	world.SOD(events, peer_session, SOD.ObjectIDs)
	h.handleANDEvent(events)
	return nil
}
