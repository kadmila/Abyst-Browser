// and (Abyss Neighbor Discovery) algorithm defines worlds in abyss network.
// AND is the algorithm provider
// A World is created by OpenWorld or JoinWorld call.
// Within a world, algorithm pushes events into the shared AND event queue.
// World algorithm may request a peer through event.
// On host-side event consumer prvides peer (by dialing or just giving connected peer).
// When a peer closes, the host calls PeerClose() to each world that references the peer.
// By this way, the host has full control over peer references.
package and

import (
	"time"

	"github.com/google/uuid"

	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
	"github.com/kadmila/Abyss-Browser/abyss_core/aurl"
	"github.com/kadmila/Abyss-Browser/abyss_core/watchdog"
)

type AND struct {
	eventCh chan IANDEvent
}

func NewAND(local_hash string) *AND {
	return &AND{
		eventCh: make(chan IANDEvent, 1024),
	}
}

func (a *AND) EventChannel() <-chan IANDEvent {
	return a.eventCh
}

func (a *AND) PeerClose(peer ani.IAbyssPeer) int {
	watchdog.Info("appCall::PeerClose " + peer.ID())

	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.stat.B(2)

	for _, world := range a.worlds {
		a.stat.B(3)
		world.RemovePeer(peer)
	}
	delete(a.peers, peer.ID())
	// TODO: drain eventCh and remove all peer-related events.
	a.eventCh <- EANDPeerClose{
		Peer: peer,
	}
	return 0
}

func (a *AND) OpenWorld(local_session_id uuid.UUID, world_url string) int {
	watchdog.Info("appCall::OpenWorld " + local_session_id.String())

	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.stat.B(4)

	world := NewWorldOpen(a, a.local_hash, local_session_id, world_url, a.peers, a.eventCh)
	a.worlds[world.lsid] = world
	return 0
}

func (a *AND) JoinWorld(local_session_id uuid.UUID, abyss_url *aurl.AURL) int {
	watchdog.Info("appCall::JoinWorld " + local_session_id.String() + " " + abyss_url.Hash)

	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.stat.B(5)

	world, err := NewWorldJoin(a, a.local_hash, local_session_id, abyss_url, a.peers, a.eventCh) //should immediate return
	if err != nil {
		return 1
	}
	a.worlds[world.lsid] = world
	return 0
}

func (a *AND) AcceptSession(local_session_id uuid.UUID, peer_session ANDPeerSession) int {
	watchdog.Info("appCall::AcceptSession " + local_session_id.String() + " " + peer_session.SessionID.String())

	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(6)
		return 0
	}
	a.stat.B(7)

	world.AcceptSession(peer_session)
	return 0
}

func (a *AND) DeclineSession(local_session_id uuid.UUID, peer_session ANDPeerSession, code int, message string) int {
	watchdog.Info("appCall::DeclineSession " + local_session_id.String() + " " + peer_session.SessionID.String())

	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(8)
		return 0
	}
	a.stat.B(9)

	world.DeclineSession(peer_session, code, message)
	return 0
}

func (a *AND) CloseWorld(local_session_id uuid.UUID) int {
	watchdog.Info("appCall::CloseWorld " + local_session_id.String())

	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(10)
		return 0
	}
	a.stat.B(11)

	world.Close()
	delete(a.worlds, local_session_id)
	return 0
}

func (a *AND) TimerExpire(local_session_id uuid.UUID) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(12)
		return 0
	}
	a.stat.B(13)

	world.TimerExpire()
	return 0
}

func (a *AND) JN(local_session_id uuid.UUID, peer_session ANDPeerSession, timestamp time.Time) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(14)
		return 0
	}
	a.stat.B(15)

	world.JN(peer_session, timestamp)
	return 0
}
func (a *AND) JOK(local_session_id uuid.UUID, peer_session ANDPeerSession, timestamp time.Time, world_url string, member_infos []ANDFullPeerSessionInfo) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(16)
		return 0
	}
	a.stat.B(17)

	world.JOK(peer_session, timestamp, world_url, member_infos)
	return 0
}
func (a *AND) JDN(local_session_id uuid.UUID, peer ani.IAbyssPeer, code int, message string) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(18)
		return 0
	}
	a.stat.B(19)

	world.JDN(peer, code, message) // after, world should be manually closed from application-side.
	return 0
}
func (a *AND) JNI(local_session_id uuid.UUID, peer_session ANDPeerSession, member_info ANDFullPeerSessionInfo) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(20)
		return 0
	}
	a.stat.B(21)

	world.JNI(peer_session, member_info)
	return 0
}
func (a *AND) MEM(local_session_id uuid.UUID, peer_session ANDPeerSession, timestamp time.Time) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(22)
		return 0
	}
	a.stat.B(23)

	world.MEM(peer_session, timestamp)
	return 0
}
func (a *AND) SJN(local_session_id uuid.UUID, peer_session ANDPeerSession, member_infos []ANDPeerSessionIdentity) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(24)
		return 0
	}
	a.stat.B(25)

	world.SJN(peer_session, member_infos)
	return 0
}
func (a *AND) CRR(local_session_id uuid.UUID, peer_session ANDPeerSession, member_infos []ANDPeerSessionIdentity) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(26)
		return 0
	}
	a.stat.B(27)

	world.CRR(peer_session, member_infos)
	return 0
}
func (a *AND) RST(local_session_id uuid.UUID, peer_session ANDPeerSession, message string) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	watchdog.Info("RST: " + message)

	if local_session_id != uuid.Nil {
		world, ok := a.worlds[local_session_id]
		if !ok {
			a.stat.B(28)
			return 0
		}
		a.stat.B(29)

		world.RST(peer_session)
	} else {
		a.stat.B(30)

		for _, world := range a.worlds {
			a.stat.B(31)
			world.RST(peer_session)
		}
	}
	return 0
}

func (a *AND) SOA(local_session_id uuid.UUID, peer_session ANDPeerSession, objects []ObjectInfo) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(32)
		return 0
	}
	a.stat.B(33)

	world.SOA(peer_session, objects)
	return 0
}
func (a *AND) SOD(local_session_id uuid.UUID, peer_session ANDPeerSession, objectIDs []uuid.UUID) int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	world, ok := a.worlds[local_session_id]
	if !ok {
		a.stat.B(34)
		return 0
	}
	a.stat.B(35)

	world.SOD(peer_session, objectIDs)
	return 0
}

func (a *AND) Statistics() string {
	return a.stat.String()
}
