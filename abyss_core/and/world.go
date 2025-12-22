package and

import (
	"math/rand"
	"net/netip"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
	"github.com/kadmila/Abyss-Browser/abyss_core/config"
	"github.com/kadmila/Abyss-Browser/abyss_core/tools/functional"
)

// World is a state machine for a world and its member/related peers.
// Removing join target from a world breakes it, so be careful.
type World struct {
	o   *AND //origin
	mtx sync.Mutex

	lsid        uuid.UUID                         // local world session id
	timestamp   time.Time                         // local world session creation timestamp
	join_target *peerWorldSessionState            // (when constructed with Join) join target peer, turns null after firing EANDWorldEnter
	join_path   string                            // (when constructed with Join) world request path
	url         string                            // (when constructed with Open, or Join accepted) environmental content URL.
	entries     map[string]*peerWorldSessionState // key: id, value: peer states
}

func newWorld_Open(events *ANDEventQueue, origin *AND, world_url string) *World {
	result := &World{
		o:           origin,
		lsid:        uuid.New(),
		timestamp:   time.Now(),
		join_target: nil,
		join_path:   "",
		url:         world_url,
		entries:     make(map[string]*peerWorldSessionState),
	}
	events.Push(&EANDWorldEnter{
		World: result,
		URL:   world_url,
	})
	events.Push(&EANDTimerRequest{
		World:    result,
		Duration: time.Millisecond * 500,
	})
	return result
}

func newWorld_Join(events *ANDEventQueue, origin *AND, target ani.IAbyssPeer, target_addrs []netip.AddrPort, path string) (*World, error) {
	result := &World{
		o:         origin,
		lsid:      uuid.New(),
		timestamp: time.Now(),
		join_target: &peerWorldSessionState{
			PeerID:            target.ID(),
			Peer:              target,
			AddressCandidates: target_addrs,
		},
		join_path: path,
		url:       "",
		entries:   make(map[string]*peerWorldSessionState),
	}
	err := result.sendJN(result.join_target)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *World) CheckSanity() {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.o == nil {
		panic("world origin nil")
	}
	if w.lsid == uuid.Nil {
		panic("world lsid nil")
	}
	if w.timestamp.Before(time.Now()) {
		panic("invalid world timestamp")
	}
	if w.join_target != nil {
		// joining
		if w.join_path == "" {
			panic("world join path nil")
		}
		if w.url != "" {
			panic("world url non-nil")
		}
		for id, entry := range w.entries {
			if id == "" {
				panic(entry.state.String() + " entry with nil id")
			}
			switch entry.state {
			case WS_DC_JNI:
				panic(entry.state.String() + " must not exist (no MEM)")
			case WS_CC:
				// expecting early MEM
				if id != entry.PeerID || id != entry.Peer.ID() {
					panic(entry.state.String() + " peer id mismatch")
				}
				if entry.SessionID != uuid.Nil {
					panic(entry.state.String() + " must have nil session id")
				}
				if !entry.TimeStamp.Equal(time.Time{}) {
					panic(entry.state.String() + " must have nil timestamp")
				}
				if entry.is_session_requested {
					panic(entry.state.String() + " is_session_requested true")
				}
			case WS_JN:
				panic("World too early to be exposed")
			case WS_RMEM_NJNI:
				// early MEM
				if id != entry.PeerID || id != entry.Peer.ID() {
					panic(entry.state.String() + " peer id mismatch")
				}
				if entry.SessionID != uuid.Nil {
					panic(entry.state.String() + " must have non-nil session id")
				}
				if entry.TimeStamp.Equal(time.Time{}) {
					panic(entry.state.String() + " must have non-nil timestamp")
				}
				if entry.is_session_requested {
					panic(entry.state.String() + " is_session_requested true")
				}
			case WS_JNI:
				panic(entry.state.String() + " must not exist (no MEM)")
			case WS_RMEM:
				panic(entry.state.String() + " must not exist (no JNI)")
			case WS_TMEM:
				panic(entry.state.String() + " must not exist (no JNI)")
			case WS_MEM:
				panic("MEM must not exist")
			default:
				panic(entry.state.String())
			}
		}
	} else {
		// working
		if w.join_path != "" {
			panic("world join path non-nil")
		}
		if w.url == "" {
			panic("world url nil")
		}
		for id, entry := range w.entries {
			if id == "" {
				panic(entry.state.String() + " entry with nil id")
			}
			switch entry.state {
			case WS_DC_JNI:
				if id != entry.PeerID {
					panic(entry.state.String() + " peer id mismatch")
				}
				if entry.Peer != nil {
					panic(entry.state.String() + " must have nil Peer")
				}
				if entry.SessionID != uuid.Nil {
					panic(entry.state.String() + " must have nil session id")
				}
				if !entry.TimeStamp.Equal(time.Time{}) {
					panic(entry.state.String() + " must have nil TimeStamp")
				}
				if entry.is_session_requested {
					panic(entry.state.String() + " is_session_requested true")
				}
			case WS_CC:
				if id != entry.PeerID || id != entry.Peer.ID() {
					panic(entry.state.String() + " peer id mismatch")
				}
				if entry.SessionID != uuid.Nil {
					panic(entry.state.String() + " must have nil session id")
				}
				if !entry.TimeStamp.Equal(time.Time{}) {
					panic(entry.state.String() + " must have nil timestamp")
				}
				if entry.is_session_requested {
					panic(entry.state.String() + " is_session_requested true")
				}
			case WS_JN, WS_JNI, WS_RMEM, WS_TMEM, WS_MEM:
				if id != entry.PeerID || id != entry.Peer.ID() {
					panic(entry.state.String() + " peer id mismatch")
				}
				if entry.SessionID == uuid.Nil {
					panic(entry.state.String() + " must have non-nil session id")
				}
				if entry.TimeStamp.Equal(time.Time{}) {
					panic(entry.state.String() + " must have non-nil timestamp")
				}
				if !entry.is_session_requested {
					panic(entry.state.String() + " is_session_requested false")
				}
			case WS_RMEM_NJNI:
				if id != entry.PeerID || id != entry.Peer.ID() {
					panic(entry.state.String() + " peer id mismatch")
				}
				if entry.SessionID == uuid.Nil {
					panic(entry.state.String() + " must have non-nil session id")
				}
				if entry.TimeStamp.Equal(time.Time{}) {
					panic(entry.state.String() + " must have non-nil timestamp")
				}
				if entry.is_session_requested {
					panic(entry.state.String() + " is_session_requested true")
				}
			default:
				panic(entry.state.String())
			}
		}
	}
}

// ContainedPeers should only be called after the world termination.
func (w *World) ContainedPeers() []ani.IAbyssPeer {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return functional.Filter_MtS(w.entries, func(s *peerWorldSessionState) ani.IAbyssPeer {
		return s.Peer
	})
}

// removeEntry should only be called for unexpected malfunction of the opponent.
// is this a good design? IDK ¯\_(ツ)_/¯
func (w *World) removeEntry(events *ANDEventQueue, entry *peerWorldSessionState, code int, message string) {
	if entry.state == WS_DC_JNI {
		delete(w.entries, entry.PeerID)
		return
	}

	if entry.state == WS_JN {
		w.sendJDN(entry, code, message)
	} else if entry.SessionID != uuid.Nil {
		w.sendRST(entry, code, message)
	}
	if entry.is_session_requested {
		events.Push(&EANDSessionClose{
			World:          w,
			ANDPeerSession: entry.ANDPeerSession(),
		})
	}
	if entry.Peer != nil {
		events.Push(&EANDPeerDiscard{
			World: w,
			Peer:  entry.Peer,
		})
	}
	delete(w.entries, entry.PeerID)
}

// tryOverwritePeerSession cleanly resets peer states if newer session id was given.
// This is kinda dangerous; impact is high. Can we ever prevent/detect forgery?
func (w *World) tryOverwritePeerSession(events *ANDEventQueue, s *peerWorldSessionState, session_id uuid.UUID, timestamp time.Time) bool {
	if s.TimeStamp.Before(timestamp) {
		switch s.state {
		case WS_DC_JNI:
			// nothing to change
		case WS_JN:
			w.sendJDN(s, JNC_OVERRUN, JNM_OVERRUN)
		default:
			w.sendRST(s, JNC_OVERRUN, JNM_OVERRUN)
		}
		if s.is_session_requested {
			events.Push(&EANDSessionClose{
				World:          w,
				ANDPeerSession: s.ANDPeerSession(),
			})
		}
		s.state = 0 // state must be defined right afterwards.
		s.SessionID = session_id
		s.TimeStamp = timestamp
		s.is_session_requested = false
		s.sjnp = false
		s.sjnc = 0
		return true
	} else {
		return false
	}
}

func (w *World) PeerConnected(events *ANDEventQueue, peer ani.IAbyssPeer, addrs []netip.AddrPort) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	config.IF_DEBUG(func() {
		if w.join_target.PeerID == peer.ID() {
			panic("duplicate peer connection")
		}
	})

	entry, ok := w.entries[peer.ID()]
	if ok { // WS_DC_JNI
		config.IF_DEBUG(func() {
			if entry.state != WS_DC_JNI {
				panic("duplicate peer connection")
			}
		})
		entry.state = WS_JNI
		entry.Peer = peer
		entry.AddressCandidates = addrs
		events.Push(&EANDSessionRequest{
			World:          w,
			ANDPeerSession: entry.ANDPeerSession(),
		})
		entry.is_session_requested = true
		return
	}

	// new entry
	w.entries[peer.ID()] = &peerWorldSessionState{
		state:             WS_CC,
		PeerID:            peer.ID(),
		Peer:              peer,
		AddressCandidates: addrs,
	}
}
func (w *World) JN(events *ANDEventQueue, peer_session ANDPeerSession, timestamp time.Time) {
	config.IF_DEBUG(func() {
		if w.join_target != nil {
			panic("JN: yet world joining") // JN is only forwarded by path - which should not be binded yet.
		}
	})

	w.mtx.Lock()
	defer w.mtx.Unlock()

	entry := w.entries[peer_session.Peer.ID()]
	config.IF_DEBUG(func() {
		if entry.Peer == nil {
			panic("JN from " + entry.state.String())
		}
	})

	if w.tryOverwritePeerSession(events, entry, peer_session.SessionID, timestamp) {
		entry.state = WS_JN
		events.Push(&EANDSessionRequest{
			World:          w,
			ANDPeerSession: peer_session,
		})
		entry.is_session_requested = true
	} else {
		w.sendJDN_Direct(peer_session, JNC_REDUNDANT, JNM_REDUNDANT)
	}
}
func (w *World) JOK(events *ANDEventQueue, peer_session ANDPeerSession, timestamp time.Time, world_url string, member_infos []ANDFullPeerSessionInfo) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	// normal case
	if w.join_target.Peer == peer_session.Peer {
		w.join_target = nil
		w.join_path = ""
		w.url = world_url

		first_member := w.join_target
		first_member.state = WS_MEM
		first_member.SessionID = peer_session.SessionID
		first_member.TimeStamp = timestamp
		first_member.is_session_requested = true
		first_member.sjnp = true

		w.entries[first_member.PeerID] = first_member

		events.Push(&EANDWorldEnter{
			World: w,
			URL:   world_url,
		})

		for _, mem_info := range member_infos {
			w.jni_mems(events, first_member, mem_info)
		}
		return
	}

	// faulty cases
	if entry, ok := w.entries[peer_session.Peer.ID()]; ok {
		w.sendRST_Direct(peer_session, JNC_INVALID_STATES, JNM_INVALID_STATES)
		w.removeEntry(events, entry, JNC_INVALID_STATES, JNM_INVALID_STATES)
		return
	}
	panic("JOK: World corrupted")
}
func (w *World) JDN(events *ANDEventQueue, peer ani.IAbyssPeer, code int, message string) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	// normal case
	if w.join_target.Peer == peer {
		events.Push(&EANDWorldLeave{
			World:   w,
			Code:    code,
			Message: message,
		})
		return
	}

	// faulty cases
	if entry, ok := w.entries[peer.ID()]; ok {
		w.removeEntry(events, entry, JNC_INVALID_STATES, JNM_INVALID_STATES)
		return
	}
	panic("JDN: World corrupted")
}
func (w *World) JNI(events *ANDEventQueue, peer_session ANDPeerSession, member_info ANDFullPeerSessionInfo) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	// only the members can send JNI.

	entry, ok := w.entries[peer_session.Peer.ID()]
	if !ok {
		// no entry exists.
		if w.join_target.Peer == peer_session.Peer {
			// join target sent JNI. Now we can't join, as joining process has beed corrupted.
			events.Push(&EANDWorldLeave{
				World:   w,
				Code:    JNC_INVALID_STATES,
				Message: JNM_INVALID_STATES,
			})
			return
		}
		// no entry, not a join target
		// a peer must be registered in before raising any communication event.
		panic("JNI: World corrupted")
	}
	// a peer has sent me JNI.
	// check if the peer is a member.
	if entry.SessionID != peer_session.SessionID {
		// session ID mismatch
		// JNI cannot overrun another session, as a membership handshake is mendated before JNI.
		// Interpretation: this is from an outdated session. Message ignored
		return
	}
	if entry.state != WS_MEM {
		// right session, wrong state - opponent failure.
		w.removeEntry(events, entry, JNC_INVALID_STATES, JNM_INVALID_STATES)
		return
	}
	// Confirmed: This JNI is from a valid member.
	w.jni_mems(events, entry, member_info)
}
func (w *World) jni_mems(events *ANDEventQueue, sender *peerWorldSessionState, mem_info ANDFullPeerSessionInfo) {
	config.IF_DEBUG(func() {
		if w.join_target != nil {
			panic("jni_mems: world is joining")
		}
	})

	mem_entry, ok := w.entries[mem_info.PeerID]
	if !ok {
		w.entries[mem_info.PeerID] = &peerWorldSessionState{
			state:     WS_DC_JNI,
			PeerID:    mem_info.PeerID,
			SessionID: mem_info.SessionID,
			TimeStamp: mem_info.TimeStamp,
		}
		events.Push(&EANDPeerRequest{
			PeerID:                     mem_info.PeerID,
			AddressCandidates:          mem_info.AddressCandidates,
			RootCertificateDer:         mem_info.RootCertificateDer,
			HandshakeKeyCertificateDer: mem_info.HandshakeKeyCertificateDer,
		})
		return
	}

	// entry exists.
	if w.tryOverwritePeerSession(events, mem_entry, mem_info.SessionID, mem_info.TimeStamp) {
		if mem_entry.Peer == nil {
			mem_entry.state = WS_DC_JNI
			events.Push(&EANDPeerRequest{
				PeerID:                     mem_info.PeerID,
				AddressCandidates:          mem_info.AddressCandidates,
				RootCertificateDer:         mem_info.RootCertificateDer,
				HandshakeKeyCertificateDer: mem_info.HandshakeKeyCertificateDer,
			})
		} else {
			mem_entry.state = WS_JNI
			events.Push(&EANDSessionRequest{
				World:          w,
				ANDPeerSession: mem_entry.ANDPeerSession(),
			})
		}
	}
}
func (w *World) MEM(events *ANDEventQueue, peer_session ANDPeerSession, timestamp time.Time) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	// MEM is onemost simple but tricky message. Any peer can send MEM, and
	// MEM can overrun old session; and it is forced, as it is from the peer.

	// only malicious case - join target sending MEM.
	if w.join_target != nil && w.join_target.Peer == peer_session.Peer {
		// join process corrupted
		events.Push(&EANDWorldLeave{
			World:   w,
			Code:    JNC_INVALID_STATES,
			Message: JNM_INVALID_STATES,
		})
		return
	}

	entry := w.entries[peer_session.Peer.ID()]
	if entry.SessionID != peer_session.SessionID {
		// MEM for unexpected session, or
		// no previous session information exists.
		if w.tryOverwritePeerSession(events, entry, peer_session.SessionID, timestamp) {
			// re-configure state, no further action can be taken.
			entry.state = WS_RMEM_NJNI
			return
		} else {
			// reset this MEM.
			w.sendRST_Direct(peer_session, JNC_OVERRUN, JNM_OVERRUN)
			return
		}
	}
	// Confirmed: This MEM is from an expected peer.

	switch entry.state {
	case WS_DC_JNI, WS_CC:
		panic("impossible")
	case WS_JN:
		// Joined and also sent MEM.
		// This is a failure, because
		// 1) joining session does not have a member.
		// 2) MEM can only be fired for JNI.
		// 3) JNI can only be sent from a member.
		// Therefore, a joining peer must not send MEM.
		w.sendRST_Direct(peer_session, JNC_INVALID_STATES, JNM_INVALID_STATES)
		w.removeEntry(events, entry, JNC_INVALID_STATES, JNM_INVALID_STATES)
		return
	case WS_RMEM_NJNI, WS_RMEM, WS_MEM:
		// very weird case - session check passed, duplicate MEM.
		// There is absolutely no need for this.
		w.removeEntry(events, entry, JNC_INVALID_STATES, JNM_INVALID_STATES)
		return
	case WS_JNI:
		entry.state = WS_RMEM
	case WS_TMEM:
		entry.state = WS_MEM
		events.Push(&EANDSessionReady{
			World:          w,
			ANDPeerSession: entry.ANDPeerSession(),
		})
	}
}
func (w *World) SJN(peer_session ANDPeerSession, member_infos []ANDPeerSessionIdentity) {
	info := w.entries[peer_session.Peer.ID()]
	if !w.isProperMemberOrReset(info, peer_session) {
		return
	}
	for _, mem_info := range member_infos {
		w.SJN_MEMS(peer_session, mem_info)
	}
}
func (w *World) SJN_MEMS(origin ANDPeerSession, mem_info ANDPeerSessionIdentity) {
	if mem_info.PeerID == w.o.local_id {
		return
	}

	info, ok := w.entries[mem_info.PeerID]
	if ok && info.state == WS_MEM && info.SessionID == mem_info.SessionID {
		info.sjnc++
		return
	}
	SendCRR(origin.Peer, w.lsid, origin.SessionID, []ANDPeerSessionIdentity{mem_info})
}
func (w *World) CRR(peer_session ANDPeerSession, member_infos []ANDPeerSessionIdentity) {
	info := w.entries[peer_session.Peer.ID()]
	if !w.isProperMemberOrReset(info, peer_session) {
		return
	}
	for _, mem_info := range member_infos {
		w.CRR_MEMS(info, mem_info)
	}
}
func (w *World) CRR_MEMS(origin *peerWorldSessionState, mem_info ANDPeerSessionIdentity) {
	if mem_info.PeerID == w.o.local_id {
		return
	}

	info, ok := w.entries[mem_info.PeerID]
	if ok && info.SessionID == mem_info.SessionID {
		SendJNI(origin.Peer, w.lsid, origin.SessionID, info.PeerWorldSession)
		SendJNI(info.Peer, w.lsid, info.SessionID, origin.PeerWorldSession)
	}
}
func (w *World) SOA(peer_session ANDPeerSession, objects []ObjectInfo) {
	info := w.entries[peer_session.Peer.ID()]
	if info.SessionID != peer_session.SessionID {
		SendRST(peer_session.Peer, w.lsid, peer_session.SessionID, "SOA::sessionID mismatch")
		return
	}
	switch info.state {
	case WS_MEM:
		w.o.eventCh <- &EANDObjectAppend{
			World:          w.lsid,
			ANDPeerSession: peer_session,
			Objects:        objects,
		}
	default:
	}
}
func (w *World) SOD(peer_session ANDPeerSession, objectIDs []uuid.UUID) {
	info := w.entries[peer_session.Peer.ID()]
	if info.SessionID != peer_session.SessionID {
		SendRST(peer_session.Peer, w.lsid, peer_session.SessionID, "SOA::sessionID mismatch")
		return
	}
	switch info.state {
	case WS_MEM:
		w.o.eventCh <- &EANDObjectDelete{
			World:          w.lsid,
			ANDPeerSession: peer_session,
			ObjectIDs:      objectIDs,
		}
	default:
	}
}
func (w *World) RST(peer_session ANDPeerSession) {
	info := w.entries[peer_session.Peer.ID()]
	w.removeEntry(info.Peer.ID(), info, "RST received")
}

func (w *World) AcceptSession(peer_session ANDPeerSession) {
	info, ok := w.entries[peer_session.Peer.ID()]
	if !ok {
		return
	}
	switch info.state {
	case WS_DC_JNI:
	case WS_CC:
		//ignore
	case WS_JT:
		panic("and invalid state: AcceptSession")
	case WS_JN:
		if info.SessionID != peer_session.SessionID {
			return
		}

		SendJOK(info.Peer, w.lsid, info.SessionID, w.timestamp, w.url, member_infos)
		info.state = WS_TMEM
	case WS_RMEM_NJNI:
		//ignore
	case WS_JNI:
		if info.SessionID != peer_session.SessionID {
			return
		}
		SendMEM(info.Peer, w.lsid, info.SessionID, w.timestamp)
		info.state = WS_TMEM
	case WS_RMEM:
		if info.SessionID != peer_session.SessionID {
			return
		}
		SendMEM(info.Peer, w.lsid, info.SessionID, w.timestamp)
		w.o.eventCh <- &EANDSessionReady{
			World:          w.lsid,
			ANDPeerSession: info.ANDPeerSession(),
		}
		info.state = WS_MEM
	case WS_TMEM:
	case WS_MEM:
		//ignore
	default:
	}
}
func (w *World) DeclineSession(peer_session ANDPeerSession, code int, message string) {
	info, ok := w.entries[peer_session.Peer.ID()]
	if !ok {
		return
	}
	if info.SessionID == peer_session.SessionID {
		//TODO: proper JDN
		w.removeEntry(peer_session.Peer.ID(), info, "application-DeclineSession called")
	}
}
func (w *World) TimerExpire() {
	sjn_mem := make([]ANDPeerSessionIdentity, 0)
	for _, info := range w.entries {
		if info.state != WS_MEM ||
			time.Since(info.TimeStamp) < time.Second ||
			info.sjnp || info.sjnc > 3 {
			continue
		}
		sjn_mem = append(sjn_mem, ANDPeerSessionIdentity{
			PeerID: info.Peer.ID(),
			World:  info.SessionID,
		})
		info.sjnc++
	}

	member_count := 0
	for _, info := range w.entries {
		if info.state != WS_MEM {
			continue
		}
		member_count++
		if len(sjn_mem) != 0 {
			SendSJN(info.Peer, w.lsid, info.SessionID, sjn_mem)
		}
	}

	w.o.eventCh <- &EANDTimerRequest{
		World:    w.lsid,
		Duration: time.Millisecond * time.Duration(300+rand.Intn(300*(member_count+1))),
	}
}

func (w *World) removeEntry(peer ani.IAbyssPeer) {
	w.removeEntry(peer.ID(), w.entries[peer.ID()], "")
	delete(w.entries, peer.ID())
}
func (w *World) Close() {
	for _, info := range w.entries {
		switch info.state {
		case WS_CC:
			//nothing
		case WS_JT:
			SendRST(info.Peer, w.lsid, info.SessionID, "Close")

			w.o.eventCh <- &EANDJoinFail{
				World:   w.lsid,
				Code:    JNC_CANCELED,
				Message: JNM_CANCELED,
			}
		case WS_JN, WS_RMEM_NJNI, WS_JNI, WS_RMEM, WS_TMEM:
			SendRST(info.Peer, w.lsid, info.SessionID, "Close")

		case WS_MEM:
			SendRST(info.Peer, w.lsid, info.SessionID, "Close")

			w.o.eventCh <- &EANDSessionClose{
				World:          w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
	}
	w.o.eventCh <- &EANDWorldLeave{
		World: w.lsid,
	}
}
