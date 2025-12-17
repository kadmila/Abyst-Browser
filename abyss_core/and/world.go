package and

import (
	"errors"
	"math/rand"
	"net/netip"
	"time"

	"github.com/google/uuid"

	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
	"github.com/kadmila/Abyss-Browser/abyss_core/aurl"
)

type ANDSessionState int

const (
	WS_DC_JNI    ANDSessionState = iota + 1 //disconnected, JNI received
	WS_CC                                   //connected, no info
	WS_JT                                   //JN sent
	WS_JN                                   //JN received
	WS_RMEM_NJNI                            //MEM received, JNI not received
	WS_JNI                                  //JNI received, MEM not received
	WS_RMEM                                 //MEM received
	WS_TMEM                                 //MEM sent
	WS_MEM                                  //member
)

// PeerWorldSessionState represents the peer's state in world session lifecycle.
// timestamp is used only for JNI.
type PeerWorldSessionState struct {
	//latest
	PeerWorldSession
	state ANDSessionState
	sjnp  bool //is sjn suppressed
	sjnc  int  //sjn receive count
}

func NewPeerWorldSessionState(peer ani.IAbyssPeer, addrs []netip.AddrPort,
	session_id uuid.UUID, timestamp time.Time, state ANDSessionState, sjnp bool, sjnc int) *PeerWorldSessionState {
	return &PeerWorldSessionState{
		PeerWorldSession: PeerWorldSession{
			PeerWithLocation: PeerWithLocation{
				Peer:              peer,
				AddressCandidates: addrs,
			},
			SessionID: session_id,
			TimeStamp: timestamp,
		},
		state: state,
		sjnp:  sjnp,
		sjnc:  sjnc,
	}
}

func (s *PeerWorldSessionState) ANDPeerSession() ANDPeerSession {
	return ANDPeerSession{
		Peer:      s.Peer,
		SessionID: s.SessionID,
	}
}

func (s *PeerWorldSessionState) Clear() {
	s.SessionID = uuid.Nil
	s.TimeStamp = time.Time{}
	if s.Peer == nil {
		panic("this peer must be removed, not cleared")
	} else {
		s.state = WS_CC
	}
	s.sjnp = false
	s.sjnc = 0
}

type World struct {
	o *AND //origin (debug purpose)

	local     string //local hash
	lsid      uuid.UUID
	timestamp time.Time
	join_hash string                            //const
	join_path string                            //const
	wurl      string                            //const
	peers     map[string]*PeerWorldSessionState //key: hash

	ech chan IANDEvent
}

func (w *World) CheckSanity() {
	jc := 0
	for peer_id, peer := range w.peers {
		if peer_id == w.local {
			panic("and sanity check failed: loopback connection")
		}

		switch peer.state {
		case WS_DC_JNI:
		case WS_CC:
		case WS_JT:
		case WS_JN:
		case WS_RMEM_NJNI:
		case WS_JNI:
		case WS_RMEM:
		case WS_TMEM:
		case WS_MEM:
		default:
			panic("and sanity check failed: non-existing state")
		}
	}
	if w.join_hash == "" && jc != 0 {
		panic("and sanity check failed: both join and open")
	}
	if jc > 1 {
		panic("and sanity check failed: multiple join targets")
	}
}

func NewWorldOpen(origin *AND, local_hash string, local_session_id uuid.UUID, world_url string, connected_members map[string]PeerWithLocation, event_ch chan IANDEvent) *World {
	result := &World{
		o:         origin,
		local:     local_hash,
		lsid:      local_session_id,
		timestamp: time.Now(),
		join_hash: "",
		join_path: "",
		wurl:      world_url,
		peers:     make(map[string]*PeerWorldSessionState),
		ech:       event_ch,
	}
	for peer_id, peer_loc := range connected_members {
		origin.stat.W(0)

		result.peers[peer_id] = &PeerWorldSessionState{
			PeerWorldSession: PeerWorldSession{
				PeerWithLocation: peer_loc,
			},
			state: WS_CC,
		}
	}
	result.ech <- EANDJoinSuccess{
		SessionID: local_session_id,
		URL:       world_url,
	}
	result.ech <- EANDTimerRequest{
		SessionID: result.lsid,
		Duration:  time.Millisecond * 500,
	}
	return result
}

func NewWorldJoin(origin *AND, local_hash string, local_session_id uuid.UUID, target *aurl.AURL, connected_members map[string]PeerWithLocation, event_ch chan IANDEvent) (*World, error) {
	result := &World{
		o:         origin,
		local:     local_hash,
		lsid:      local_session_id,
		timestamp: time.Now(),
		join_hash: target.Hash,
		join_path: target.Path,
		peers:     make(map[string]*PeerWorldSessionState),
		ech:       event_ch,
	}
	for peer_id, peer_loc := range connected_members {
		origin.stat.W(1)

		result.peers[peer_id] = &PeerWorldSessionState{
			PeerWorldSession: PeerWorldSession{
				PeerWithLocation: peer_loc,
			},
			state: WS_CC,
		}
	}

	if connected_target, ok := result.peers[target.Hash]; ok {
		origin.stat.W(2)

		connected_target.state = WS_JT
		origin.stat.JN_TX++
		SendJN(connected_target.Peer, local_session_id, target.Path, result.timestamp)
	} else {
		return nil, errors.New("peer not found")
	}
	return result, nil
}

func (w *World) ClearStates(peer_id string, info *PeerWorldSessionState, message string) {
	switch info.state {
	case WS_DC_JNI:
		delete(w.peers, peer_id)
	case WS_CC:
		info.Clear()
	case WS_JT:
		w.o.stat.RST_TX++
		SendRST(info.Peer, w.lsid, info.SessionID, "ClearStates::WS_JT "+message)
		w.ech <- EANDJoinFail{
			SessionID: w.lsid,
			Code:      JNC_INVALID_STATES,
			Message:   JNM_INVALID_STATES,
		}
		info.Clear()
	case WS_JN:
		w.o.stat.JDN_TX++
		SendJDN(info.Peer, info.SessionID, JNC_INVALID_STATES, JNM_INVALID_STATES)
		info.Clear()
	case WS_MEM:
		w.ech <- EANDSessionClose{
			SessionID:      w.lsid,
			ANDPeerSession: info.ANDPeerSession(),
		}
		fallthrough
	case WS_RMEM_NJNI, WS_JNI, WS_RMEM, WS_TMEM:
		w.o.stat.RST_TX++
		SendRST(info.Peer, w.lsid, info.SessionID, "ClearStates::else "+message)
		info.Clear()
	}
}

// TryUpdateSessionID returns (old session ID, success). old session ID is nil if not updated
func (w *World) TryUpdateSessionID(s *PeerWorldSessionState, session_id uuid.UUID, timestamp time.Time) bool {
	if s.TimeStamp.Before(timestamp) {
		w.ClearStates(s.Peer.ID(), s, "session id update failure")
		s.SessionID = session_id
		s.TimeStamp = timestamp
		return true
	} else {
		return false
	}
}
func (w *World) IsProperMemberOrReset(info *PeerWorldSessionState, peer_session ANDPeerSession) bool {
	switch info.state {
	case WS_DC_JNI:
		panic("not connected")
	case WS_MEM:
		if info.SessionID == peer_session.SessionID {
			return true
		}
		fallthrough
	default:
		w.ClearStates(info.Peer.ID(), info, "non-member reset")
	}
	return false
}

func (w *World) PeerConnected(peer_loc PeerWithLocation) {
	info, ok := w.peers[peer_loc.Peer.ID()]
	if ok { // known peer
		w.o.stat.W(4)

		switch info.state {
		case WS_DC_JNI:
			w.o.stat.W(6)

			info.PeerWithLocation = peer_loc
			info.state = WS_JNI

			w.ech <- EANDSessionRequest{
				SessionID:      w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		default:
			panic("and: duplicate connection")
		}

		return
	}
	//unknown peer
	w.peers[peer_loc.Peer.ID()] = &PeerWorldSessionState{
		PeerWorldSession: PeerWorldSession{
			PeerWithLocation: peer_loc,
		},
		state: WS_CC,
	}
}
func (w *World) JN(peer_session ANDPeerSession, timestamp time.Time) {
	w.o.stat.JN_RX++

	info := w.peers[peer_session.Peer.ID()]
	switch info.state {
	case WS_CC:
		w.o.stat.W(7)

		info.SessionID = peer_session.SessionID
		info.TimeStamp = timestamp
		info.state = WS_JN
		w.ech <- EANDSessionRequest{
			SessionID:      w.lsid,
			ANDPeerSession: peer_session,
		}
	case WS_JT: //should not happen. during joining, the world must be hidden, not accepting JN.
		w.o.stat.W(8)

		w.o.stat.JDN_TX++
		SendJDN(peer_session.Peer, peer_session.SessionID, JNC_INVALID_STATES, JNM_INVALID_STATES)
	case WS_JN, WS_RMEM_NJNI, WS_JNI, WS_RMEM, WS_TMEM, WS_MEM:
		w.o.stat.W(9)

		if w.TryUpdateSessionID(info, peer_session.SessionID, timestamp) {
			w.o.stat.W(10)

			info.state = WS_JN
			w.ech <- EANDSessionRequest{
				SessionID:      w.lsid,
				ANDPeerSession: peer_session,
			}
		} else {
			w.o.stat.W(11)

			w.o.stat.JDN_TX++
			SendJDN(peer_session.Peer, peer_session.SessionID, JNC_DUPLICATE, JNM_DUPLICATE) //must not happen
		}
	default:
		panic("and invalid state: JN")
	}
}
func (w *World) JOK(peer_session ANDPeerSession, timestamp time.Time, world_url string, member_infos []ANDFullPeerSessionInfo) {
	w.o.stat.JOK_RX++

	sender_id := peer_session.Peer.ID()
	info := w.peers[sender_id]
	if w.join_hash != sender_id ||
		info.state != WS_JT {
		w.o.stat.W(12)

		w.o.stat.RST_TX++
		SendRST(peer_session.Peer, w.lsid, peer_session.SessionID, "JOK::not WS_JT")
		return
	}

	w.o.stat.W(13)

	info.SessionID = peer_session.SessionID
	info.TimeStamp = timestamp
	w.ech <- EANDJoinSuccess{
		SessionID: w.lsid,
		URL:       world_url,
	}
	w.ech <- EANDSessionRequest{
		SessionID:      w.lsid,
		ANDPeerSession: peer_session,
	}
	info.state = WS_RMEM
	info.sjnp = true

	for _, mem_info := range member_infos {
		w.o.stat.W(14)

		w.JNI_MEMS(sender_id, mem_info)
	}
}
func (w *World) JDN(peer ani.IAbyssPeer, code int, message string) { //no branch number here... :(
	w.o.stat.JDN_RX++

	info := w.peers[peer.ID()]
	if w.join_hash != peer.ID() ||
		info.state != WS_JT {
		w.o.stat.W(15)

		return
	}

	w.o.stat.W(16)

	w.ech <- EANDJoinFail{
		SessionID: w.lsid,
		Code:      code,
		Message:   message,
	}
	info.Clear()
}

func (w *World) JNI(peer_session ANDPeerSession, member_info ANDFullPeerSessionInfo) {
	w.o.stat.JNI_RX++

	sender_id := peer_session.Peer.ID()
	info := w.peers[sender_id]

	if !w.IsProperMemberOrReset(info, peer_session) {
		w.o.stat.W(17)

		return
	}

	w.o.stat.W(18)

	w.JNI_MEMS(sender_id, member_info)
}
func (w *World) JNI_MEMS(sender_id string, mem_info ANDFullPeerSessionInfo) {
	peer_id := mem_info.PeerID
	if peer_id == w.local {
		w.o.stat.W(19)
		return
	}

	info, ok := w.peers[peer_id]
	if !ok {
		w.o.stat.W(20)

		w.peers[peer_id] = &PeerWorldSessionState{
			PeerWorldSession: PeerWorldSession{
				SessionID: mem_info.SessionID,
				TimeStamp: mem_info.TimeStamp,
			},
			state: WS_DC_JNI,
		}
		w.ech <- EANDConnectRequest{
			PeerID:                     mem_info.PeerID,
			AddressCandidates:          mem_info.AddressCandidates,
			RootCertificateDer:         mem_info.RootCertificateDer,
			HandshakeKeyCertificateDer: mem_info.HandshakeKeyCertificateDer,
		}
		return
	}

	switch info.state {
	case WS_JT:
		panic("and: proper member check failed (JNI)")
	case WS_DC_JNI:
		w.o.stat.W(21)

		if info.TimeStamp.Before(mem_info.TimeStamp) {
			info.SessionID = mem_info.SessionID
			info.TimeStamp = mem_info.TimeStamp
			info.state = WS_DC_JNI
		}
		//previously, tried connecting. may need to refresh connection trials
	case WS_CC:
		w.o.stat.W(22)

		info.SessionID = mem_info.SessionID
		info.TimeStamp = mem_info.TimeStamp
		info.state = WS_JNI
		w.ech <- EANDSessionRequest{
			SessionID:      w.lsid,
			ANDPeerSession: info.ANDPeerSession(),
		}
	case WS_JN:
		w.o.stat.W(23)

		if w.TryUpdateSessionID(info, mem_info.SessionID, mem_info.TimeStamp) {
			//unlikely to happen
			info.state = WS_JNI
			w.ech <- EANDSessionRequest{
				SessionID:      w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
	case WS_RMEM_NJNI:
		w.o.stat.W(24)

		if w.TryUpdateSessionID(info, mem_info.SessionID, mem_info.TimeStamp) {
			w.o.stat.W(25)

			info.state = WS_JNI
			w.ech <- EANDSessionRequest{
				SessionID:      w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
			return
		}
		if info.SessionID == mem_info.SessionID {
			w.o.stat.W(26)

			info.state = WS_RMEM
			w.ech <- EANDSessionRequest{
				SessionID:      w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
		//else: old session
	case WS_JNI, WS_RMEM, WS_TMEM, WS_MEM:
		if w.TryUpdateSessionID(info, mem_info.SessionID, mem_info.TimeStamp) {
			w.o.stat.W(27)

			info.state = WS_JNI
			w.ech <- EANDSessionRequest{
				SessionID:      w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
			return
		}
		w.o.stat.W(28)

	default:
		panic("and invalid state: JNI_MEMS")
	}
}
func (w *World) MEM(peer_session ANDPeerSession, timestamp time.Time) {
	w.o.stat.MEM_RX++

	info := w.peers[peer_session.Peer.ID()]
	switch info.state {
	case WS_CC:
		w.o.stat.W(29)

		info.SessionID = peer_session.SessionID
		info.TimeStamp = timestamp
		info.state = WS_RMEM_NJNI
	case WS_JT:
		w.o.stat.W(30)

		w.ClearStates(peer_session.Peer.ID(), info, "received MEM from WS_JT")
	case WS_JN, WS_RMEM_NJNI, WS_RMEM, WS_MEM:
		if w.TryUpdateSessionID(info, peer_session.SessionID, timestamp) {
			w.o.stat.W(31)

			info.state = WS_RMEM_NJNI
			return
		}
		w.o.stat.W(32)

	case WS_JNI:
		if w.TryUpdateSessionID(info, peer_session.SessionID, timestamp) {
			w.o.stat.W(33)

			info.state = WS_RMEM_NJNI
			return
		}
		if info.SessionID == peer_session.SessionID {
			w.o.stat.W(34)

			info.state = WS_RMEM
		}
		w.o.stat.W(35)

	case WS_TMEM:
		w.o.stat.W(36)

		if w.TryUpdateSessionID(info, peer_session.SessionID, timestamp) {
			w.o.stat.W(37)

			info.state = WS_RMEM_NJNI
			return
		}
		if info.SessionID == peer_session.SessionID {
			w.o.stat.W(38)

			info.state = WS_MEM
			w.ech <- EANDSessionReady{
				SessionID:      w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
		w.o.stat.W(39)

	default:
		panic("and: impossible disconnected state")
	}
}
func (w *World) SJN(peer_session ANDPeerSession, member_infos []ANDPeerSessionIdentity) {
	w.o.stat.SJN_RX++

	info := w.peers[peer_session.Peer.ID()]
	if !w.IsProperMemberOrReset(info, peer_session) {
		w.o.stat.W(40)

		return
	}
	for _, mem_info := range member_infos {
		w.o.stat.W(41)

		w.SJN_MEMS(peer_session, mem_info)
	}
}
func (w *World) SJN_MEMS(origin ANDPeerSession, mem_info ANDPeerSessionIdentity) {
	if mem_info.PeerID == w.local {
		w.o.stat.W(42)
		return
	}

	info, ok := w.peers[mem_info.PeerID]
	if ok && info.state == WS_MEM && info.SessionID == mem_info.SessionID {
		w.o.stat.W(43)

		info.sjnc++
		return
	}
	w.o.stat.CRR_TX++
	SendCRR(origin.Peer, w.lsid, origin.SessionID, []ANDPeerSessionIdentity{mem_info})
}
func (w *World) CRR(peer_session ANDPeerSession, member_infos []ANDPeerSessionIdentity) {
	w.o.stat.CRR_RX++

	info := w.peers[peer_session.Peer.ID()]
	if !w.IsProperMemberOrReset(info, peer_session) {
		w.o.stat.W(44)

		return
	}
	for _, mem_info := range member_infos {
		w.o.stat.W(45)

		w.CRR_MEMS(info, mem_info)
	}
}
func (w *World) CRR_MEMS(origin *PeerWorldSessionState, mem_info ANDPeerSessionIdentity) {
	if mem_info.PeerID == w.local {
		w.o.stat.W(46)
		return
	}

	info, ok := w.peers[mem_info.PeerID]
	if ok && info.SessionID == mem_info.SessionID {
		w.o.stat.W(47)

		w.o.stat.JNI_TX++
		SendJNI(origin.Peer, w.lsid, origin.SessionID, info.PeerWorldSession)
		w.o.stat.JNI_TX++
		SendJNI(info.Peer, w.lsid, info.SessionID, origin.PeerWorldSession)
	}
}
func (w *World) SOA(peer_session ANDPeerSession, objects []ObjectInfo) {
	w.o.stat.SOA_RX++

	info := w.peers[peer_session.Peer.ID()]
	if info.SessionID != peer_session.SessionID {
		w.o.stat.W(48)

		w.o.stat.RST_TX++
		SendRST(peer_session.Peer, w.lsid, peer_session.SessionID, "SOA::sessionID mismatch")
		return
	}
	switch info.state {
	case WS_MEM:
		w.o.stat.W(49)

		w.ech <- EANDObjectAppend{
			SessionID:      w.lsid,
			ANDPeerSession: peer_session,
			Objects:        objects,
		}
	default:
		w.o.stat.W(50)
	}
}
func (w *World) SOD(peer_session ANDPeerSession, objectIDs []uuid.UUID) {
	w.o.stat.SOD_RX++

	info := w.peers[peer_session.Peer.ID()]
	if info.SessionID != peer_session.SessionID {
		w.o.stat.W(51)

		w.o.stat.RST_TX++
		SendRST(peer_session.Peer, w.lsid, peer_session.SessionID, "SOA::sessionID mismatch")
		return
	}
	switch info.state {
	case WS_MEM:
		w.o.stat.W(52)

		w.ech <- EANDObjectDelete{
			SessionID:      w.lsid,
			ANDPeerSession: peer_session,
			ObjectIDs:      objectIDs,
		}
	default:
		w.o.stat.W(53)
	}
}
func (w *World) RST(peer_session ANDPeerSession) {
	w.o.stat.RST_RX++

	info := w.peers[peer_session.Peer.ID()]
	w.ClearStates(info.Peer.ID(), info, "RST received")
}

func (w *World) AcceptSession(peer_session ANDPeerSession) {
	info, ok := w.peers[peer_session.Peer.ID()]
	if !ok {
		w.o.stat.W(54)
		return
	}
	switch info.state {
	case WS_DC_JNI:
		w.o.stat.W(55)

	case WS_CC:
		w.o.stat.W(56)

		//ignore
	case WS_JT:
		panic("and invalid state: AcceptSession")
	case WS_JN:
		w.o.stat.W(57)

		if info.SessionID != peer_session.SessionID {
			w.o.stat.W(58)

			return
		}

		member_infos := make([]PeerWorldSession, 0)
		for _, p := range w.peers {
			if p.state != WS_MEM {
				w.o.stat.W(59)

				continue
			}
			w.o.stat.W(60)

			member_infos = append(member_infos, PeerWorldSession{
				PeerWithLocation: PeerWithLocation{
					Peer:              p.Peer,
					AddressCandidates: p.AddressCandidates,
				},
				SessionID: p.SessionID,
				TimeStamp: p.TimeStamp,
			})
			w.o.stat.JNI_TX++
			SendJNI(p.Peer, w.lsid, p.SessionID, info.PeerWorldSession)
		}
		w.o.stat.JOK_TX++
		SendJOK(info.Peer, w.lsid, info.SessionID, w.timestamp, w.wurl, member_infos)
		info.state = WS_TMEM
	case WS_RMEM_NJNI:
		w.o.stat.W(61)

		//ignore
	case WS_JNI:
		w.o.stat.W(62)

		if info.SessionID != peer_session.SessionID {
			w.o.stat.W(63)

			return
		}
		w.o.stat.W(64)

		w.o.stat.MEM_TX++
		SendMEM(info.Peer, w.lsid, info.SessionID, w.timestamp)
		info.state = WS_TMEM
	case WS_RMEM:
		w.o.stat.W(65)

		if info.SessionID != peer_session.SessionID {
			w.o.stat.W(66)

			return
		}
		w.o.stat.W(67)

		w.o.stat.MEM_TX++
		SendMEM(info.Peer, w.lsid, info.SessionID, w.timestamp)
		w.ech <- EANDSessionReady{
			SessionID:      w.lsid,
			ANDPeerSession: info.ANDPeerSession(),
		}
		info.state = WS_MEM
	case WS_TMEM:
		w.o.stat.W(68)

		//ignore
	case WS_MEM:
		w.o.stat.W(69)

		//ignore
	default:
		w.o.stat.W(70)
	}
}
func (w *World) DeclineSession(peer_session ANDPeerSession, code int, message string) {
	info, ok := w.peers[peer_session.Peer.ID()]
	if !ok {
		w.o.stat.W(71)
		return
	}
	if info.SessionID == peer_session.SessionID {
		w.o.stat.W(72)

		//TODO: proper JDN
		w.ClearStates(peer_session.Peer.ID(), info, "application-DeclineSession called")
	}
	w.o.stat.W(73)

}
func (w *World) TimerExpire() {
	sjn_mem := make([]ANDPeerSessionIdentity, 0)
	for _, info := range w.peers {
		if info.state != WS_MEM ||
			time.Since(info.TimeStamp) < time.Second ||
			info.sjnp || info.sjnc > 3 {
			w.o.stat.W(74)

			continue
		}
		w.o.stat.W(75)

		sjn_mem = append(sjn_mem, ANDPeerSessionIdentity{
			PeerID:    info.Peer.ID(),
			SessionID: info.SessionID,
		})
		info.sjnc++
	}

	member_count := 0
	for _, info := range w.peers {
		if info.state != WS_MEM {
			w.o.stat.W(76)

			continue
		}
		member_count++
		if len(sjn_mem) != 0 {
			w.o.stat.W(77)

			w.o.stat.SJN_TX++
			SendSJN(info.Peer, w.lsid, info.SessionID, sjn_mem)
		}
	}

	w.ech <- EANDTimerRequest{
		SessionID: w.lsid,
		Duration:  time.Millisecond * time.Duration(300+rand.Intn(300*(member_count+1))),
	}
}

func (w *World) RemovePeer(peer ani.IAbyssPeer) {
	w.ClearStates(peer.ID(), w.peers[peer.ID()], "")
	delete(w.peers, peer.ID())
}
func (w *World) Close() {
	for _, info := range w.peers {
		switch info.state {
		case WS_CC:
			//nothing
		case WS_JT:
			w.o.stat.W(78)

			w.o.stat.RST_TX++
			SendRST(info.Peer, w.lsid, info.SessionID, "Close")

			w.ech <- EANDJoinFail{
				SessionID: w.lsid,
				Code:      JNC_CANCELED,
				Message:   JNM_CANCELED,
			}
		case WS_JN, WS_RMEM_NJNI, WS_JNI, WS_RMEM, WS_TMEM:
			w.o.stat.W(79)

			w.o.stat.RST_TX++
			SendRST(info.Peer, w.lsid, info.SessionID, "Close")

		case WS_MEM:
			w.o.stat.W(80)

			w.o.stat.RST_TX++
			SendRST(info.Peer, w.lsid, info.SessionID, "Close")

			w.ech <- EANDSessionClose{
				SessionID:      w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
	}
	w.o.stat.W(81)

	w.ech <- EANDWorldLeave{
		SessionID: w.lsid,
	}
}
