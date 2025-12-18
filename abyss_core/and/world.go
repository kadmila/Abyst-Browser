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

type worldLifecycle int

const (
	W_Joining worldLifecycle = iota
	W_Working
	W_Closed
)

// World is a state machine for a world and its member/related peers.
// Removing join target from a world breakes it, so be careful.
type World struct {
	o   *AND //origin (debug purpose)
	mtx sync.Mutex

	lsid        uuid.UUID                         // local world session id
	timestamp   time.Time                         // local world session creation timestamp
	join_target *peerWorldSessionState            // (when constructed with Join) join target peer, turns null after firing EANDWorldEnter
	join_path   string                            // (when constructed with Join) world request path
	url         string                            // (when constructed with Open, or Join accepted) environmental content URL.
	entries     map[string]*peerWorldSessionState // key: id, value: peer states
	lifecycle   worldLifecycle                    // must be initialized with true, becomes false after firing EANDWorldLeave
}

func newWorld_Open(origin *AND, world_url string) *World {
	result := &World{
		o:           origin,
		lsid:        uuid.New(),
		timestamp:   time.Now(),
		join_target: nil,
		join_path:   "",
		url:         world_url,
		entries:     make(map[string]*peerWorldSessionState),
		lifecycle:   W_Working,
	}
	result.o.eventCh <- &EANDWorldEnter{
		World: result,
		URL:   world_url,
	}
	result.o.eventCh <- &EANDTimerRequest{
		World:    result,
		Duration: time.Millisecond * 500,
	}
	return result
}

func newWorld_Join(origin *AND, target ani.IAbyssPeer, target_addrs []netip.AddrPort, path string) (*World, error) {
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
		lifecycle: W_Joining,
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
	switch w.lifecycle {
	case W_Joining:
		if w.join_target == nil {
			panic("world join target nil")
		}
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
	case W_Working:
		if w.join_target != nil {
			panic("world join target non-nil")
		}
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
	case W_Closed:
		if w.join_target != nil {
			panic("world join target non-nil")
		}
		if w.join_path != "" {
			panic("world join path non-nil")
		}
		if w.url == "" {
			panic("world url nil")
		}
		for id, entry := range w.entries {
			if id == "" {
				panic("entry with nil id")
			}
			switch entry.state {
			case WS_DC_JNI:
			case WS_CC:
			case WS_JN:
			case WS_RMEM_NJNI:
			case WS_JNI:
			case WS_RMEM:
			case WS_TMEM:
			case WS_MEM:
			}
		}
	default:
		panic("invalid world lifecycle")
	}
}

// ContainedPeers should only be called after the world termination.
func (w *World) ContainedPeers() []ani.IAbyssPeer {
	config.IF_DEBUG(func() {
		if w.lifecycle != W_Closed {
			panic("World.ContainedPeers should only be used after the world closes.")
		}
	})

	w.mtx.Lock()
	defer w.mtx.Unlock()

	return functional.Filter_MtS(w.entries, func(s *peerWorldSessionState) ani.IAbyssPeer {
		return s.Peer
	})
}

// removeEntry should only be called for unexpected malfunction of the opponent.
// is this a good design? IDK ¯\_(ツ)_/¯
func (w *World) removeEntry(entry *peerWorldSessionState, message string) {
	config.IF_DEBUG(func() {
		if w.lifecycle != W_Working {
			panic("World.removeEntry should only be used when the world is working state.")
		}
	})

	switch entry.state {
	case WS_DC_JNI:

	}

	if entry.state == WS_JN {
		w.sendJDN(entry, JNC_INVALID_STATES, JNM_INVALID_STATES)
	}
	if entry.SessionID != uuid.Nil {
		w.sendRST(entry, "unexpected failure::"+entry.state.String()+":"+message)
	}
	if entry.is_session_requested {
		w.o.eventCh <- &EANDSessionClose{
			World:          w,
			ANDPeerSession: entry.ANDPeerSession(),
		}
	}
	if entry.Peer != nil {
		w.o.eventCh <- &EANDPeerDiscard{
			World: w,
			Peer:  entry.Peer,
		}
	}
	delete(w.entries, entry.PeerID)
}

// tryOverwritePeerSession cleanly resets peer states if newer session id was given.
// is this a good design? IDK ¯\_(ツ)_/¯
func (w *World) tryOverwritePeerSession(s *peerWorldSessionState, session_id uuid.UUID, timestamp time.Time) bool {
	if s.TimeStamp.Before(timestamp) {
		s.state = 0 // state must be defined right afterwards.
		s.SessionID = session_id
		s.TimeStamp = timestamp
		if s.is_session_requested {
			w.o.eventCh <- EANDSessionClose{
				World:          w,
				ANDPeerSession: s.ANDPeerSession(),
			}
		}
		s.is_session_requested = false
		s.sjnp = false
		s.sjnc = 0
		return true
	} else {
		return false
	}
}

// is this a good design? IDK ¯\_(ツ)_/¯
func (w *World) isProperMemberOrReset(info *peerWorldSessionState, peer_session ANDPeerSession) bool {
	switch info.state {
	case WS_DC_JNI:
		panic("not connected")
	case WS_MEM:
		if info.SessionID == peer_session.SessionID {
			return true
		}
		fallthrough
	default:
		w.removeEntry(info, "non-member reset")
	}
	return false
}

func (w *World) PeerConnected(peer_loc PeerWithLocation) {
	info, ok := w.entries[peer_loc.Peer.ID()]
	if ok { // known peer
		switch info.state {
		case WS_DC_JNI:
			info.PeerWithLocation = peer_loc
			info.state = WS_JNI

			w.o.eventCh <- &EANDSessionRequest{
				World:          w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		default:
			panic("and: duplicate connection")
		}

		return
	}
	//unknown peer
	w.entries[peer_loc.Peer.ID()] = &peerWorldSessionState{
		PeerWorldSession: PeerWorldSession{
			PeerWithLocation: peer_loc,
		},
		state: WS_CC,
	}
}
func (w *World) JN(peer_session ANDPeerSession, timestamp time.Time) {
	info := w.entries[peer_session.Peer.ID()]
	switch info.state {
	case WS_CC:
		info.SessionID = peer_session.SessionID
		info.TimeStamp = timestamp
		info.state = WS_JN
		w.o.eventCh <- &EANDSessionRequest{
			World:          w.lsid,
			ANDPeerSession: peer_session,
		}
	case WS_JT: //should not happen. during joining, the world must be hidden, not accepting JN.
		SendJDN(peer_session.Peer, peer_session.SessionID, JNC_INVALID_STATES, JNM_INVALID_STATES)
	case WS_JN, WS_RMEM_NJNI, WS_JNI, WS_RMEM, WS_TMEM, WS_MEM:
		if w.tryOverwritePeerSession(info, peer_session.SessionID, timestamp) {
			info.state = WS_JN
			w.o.eventCh <- &EANDSessionRequest{
				World:          w.lsid,
				ANDPeerSession: peer_session,
			}
		} else {
			SendJDN(peer_session.Peer, peer_session.SessionID, JNC_DUPLICATE, JNM_DUPLICATE) //must not happen
		}
	default:
		panic("and invalid state: JN")
	}
}
func (w *World) JOK(peer_session ANDPeerSession, timestamp time.Time, world_url string, member_infos []ANDFullPeerSessionInfo) {
	sender_id := peer_session.Peer.ID()
	info := w.entries[sender_id]
	if w.join_target != sender_id ||
		info.state != WS_JT {
		SendRST(peer_session.Peer, w.lsid, peer_session.SessionID, "JOK::not WS_JT")
		return
	}

	info.SessionID = peer_session.SessionID
	info.TimeStamp = timestamp
	w.o.eventCh <- &EANDWorldEnter{
		World: w.lsid,
		URL:   world_url,
	}
	w.o.eventCh <- &EANDSessionRequest{
		World:          w.lsid,
		ANDPeerSession: peer_session,
	}
	info.state = WS_RMEM
	info.sjnp = true

	for _, mem_info := range member_infos {
		w.JNI_MEMS(sender_id, mem_info)
	}
}
func (w *World) JDN(peer ani.IAbyssPeer, code int, message string) { //no branch number here... :(
	info := w.entries[peer.ID()]
	if w.join_target != peer.ID() ||
		info.state != WS_JT {
		return
	}

	w.o.eventCh <- &EANDJoinFail{
		World:   w.lsid,
		Code:    code,
		Message: message,
	}
	info.Clear()
}

func (w *World) JNI(peer_session ANDPeerSession, member_info ANDFullPeerSessionInfo) {
	sender_id := peer_session.Peer.ID()
	info := w.entries[sender_id]

	if !w.isProperMemberOrReset(info, peer_session) {
		return
	}

	w.JNI_MEMS(sender_id, member_info)
}
func (w *World) JNI_MEMS(sender_id string, mem_info ANDFullPeerSessionInfo) {
	peer_id := mem_info.PeerID
	if peer_id == w.o.local_id {
		return
	}

	info, ok := w.entries[peer_id]
	if !ok {
		w.entries[peer_id] = &peerWorldSessionState{
			PeerWorldSession: PeerWorldSession{
				World:     mem_info.SessionID,
				TimeStamp: mem_info.TimeStamp,
			},
			state: WS_DC_JNI,
		}
		w.o.eventCh <- &EANDPeerRequest{
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
		if info.TimeStamp.Before(mem_info.TimeStamp) {
			info.SessionID = mem_info.SessionID
			info.TimeStamp = mem_info.TimeStamp
			info.state = WS_DC_JNI
		}
		//previously, tried connecting. may need to refresh connection trials
	case WS_CC:
		info.SessionID = mem_info.SessionID
		info.TimeStamp = mem_info.TimeStamp
		info.state = WS_JNI
		w.o.eventCh <- &EANDSessionRequest{
			World:          w.lsid,
			ANDPeerSession: info.ANDPeerSession(),
		}
	case WS_JN:
		if w.tryOverwritePeerSession(info, mem_info.SessionID, mem_info.TimeStamp) {
			//unlikely to happen
			info.state = WS_JNI
			w.o.eventCh <- &EANDSessionRequest{
				World:          w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
	case WS_RMEM_NJNI:
		if w.tryOverwritePeerSession(info, mem_info.SessionID, mem_info.TimeStamp) {
			info.state = WS_JNI
			w.o.eventCh <- &EANDSessionRequest{
				World:          w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
			return
		}
		if info.SessionID == mem_info.SessionID {
			info.state = WS_RMEM
			w.o.eventCh <- &EANDSessionRequest{
				World:          w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
		//else: old session
	case WS_JNI, WS_RMEM, WS_TMEM, WS_MEM:
		if w.tryOverwritePeerSession(info, mem_info.SessionID, mem_info.TimeStamp) {
			info.state = WS_JNI
			w.o.eventCh <- &EANDSessionRequest{
				World:          w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
			return
		}
	default:
		panic("and invalid state: JNI_MEMS")
	}
}
func (w *World) MEM(peer_session ANDPeerSession, timestamp time.Time) {
	info := w.entries[peer_session.Peer.ID()]
	switch info.state {
	case WS_CC:
		info.SessionID = peer_session.SessionID
		info.TimeStamp = timestamp
		info.state = WS_RMEM_NJNI
	case WS_JT:
		w.removeEntry(peer_session.Peer.ID(), info, "received MEM from WS_JT")
	case WS_JN, WS_RMEM_NJNI, WS_RMEM, WS_MEM:
		if w.tryOverwritePeerSession(info, peer_session.SessionID, timestamp) {
			info.state = WS_RMEM_NJNI
			return
		}
	case WS_JNI:
		if w.tryOverwritePeerSession(info, peer_session.SessionID, timestamp) {
			info.state = WS_RMEM_NJNI
			return
		}
		if info.SessionID == peer_session.SessionID {
			info.state = WS_RMEM
		}
	case WS_TMEM:
		if w.tryOverwritePeerSession(info, peer_session.SessionID, timestamp) {
			info.state = WS_RMEM_NJNI
			return
		}
		if info.SessionID == peer_session.SessionID {
			info.state = WS_MEM
			w.o.eventCh <- &EANDSessionReady{
				World:          w.lsid,
				ANDPeerSession: info.ANDPeerSession(),
			}
		}
	default:
		panic("and: impossible disconnected state")
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
