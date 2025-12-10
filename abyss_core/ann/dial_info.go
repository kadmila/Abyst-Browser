package ann

import (
	"context"
	"net/netip"
	"sync"

	"slices"

	"github.com/kadmila/Abyss-Browser/abyss_core/sec"
	"github.com/kadmila/Abyss-Browser/abyss_core/tools/waiter"
)

type DialErrorType int

const (
	DE_Redundant DialErrorType = iota + 1
	DE_UnknownPeer
)

type DialError struct {
	T DialErrorType
}

func (e *DialError) Error() string {
	switch e.T {
	case DE_Redundant:
		return "redundant"
	case DE_UnknownPeer:
		return "unknown peer"
	default:
		return "fatal::Memory Corruption"
	}
}

// DialInfoMap provides peer identity and helps avoiding
// redundant dials, but it may still allow some redundancy.
type DialInfoMap struct {
	mtx sync.Mutex

	known   map[string]*sec.AbyssPeerIdentity
	waiting map[string]*waiter.Waiter[*sec.AbyssPeerIdentity]
	dialed  map[string][]netip.Addr
}

func MakeDialInfoMap() DialInfoMap {
	return DialInfoMap{
		known:   make(map[string]*sec.AbyssPeerIdentity),
		waiting: make(map[string]*waiter.Waiter[*sec.AbyssPeerIdentity]),
		dialed:  make(map[string][]netip.Addr),
	}
}

func (m *DialInfoMap) UpdatePeerInformation(identity *sec.AbyssPeerIdentity) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	old_identity, ok := m.known[identity.ID()]
	if ok && old_identity.IssueTime().After(identity.IssueTime()) {
		return
	}

	m.known[identity.ID()] = identity
	if waiter, ok := m.waiting[identity.ID()]; ok {
		waiter.Set(identity)
		delete(m.waiting, identity.ID())
	}

	// peer information updated - new handshake key, all old ongoing dials and connections will fail.
	delete(m.dialed, identity.ID())
}

func (m *DialInfoMap) Remove(id string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	delete(m.known, id)
}

func (m *DialInfoMap) CleaupWaiter() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for id, waiter := range m.waiting {
		if _, ok := waiter.TryClose(); ok {
			delete(m.waiting, id)
		}
	}
}

func (m *DialInfoMap) Get(ctx context.Context, id string) (*sec.AbyssPeerIdentity, error) {
	var identity *sec.AbyssPeerIdentity
	var is_known bool
	var identity_waiter *waiter.Waiter[*sec.AbyssPeerIdentity]

	for {
		// get if known. If unknown, get or register waiter.
		m.mtx.Lock()
		identity, is_known = m.known[id]
		if !is_known {
			var waiter_found bool
			identity_waiter, waiter_found = m.waiting[id]
			if !waiter_found {
				identity_waiter = waiter.NewWaiter[*sec.AbyssPeerIdentity]()
				m.waiting[id] = identity_waiter
			}
		}
		m.mtx.Unlock()

		// known identity found.
		if is_known {
			return identity, nil
		}

		// waiter found or registered. Try waiting.
		var wait_err error
		identity, wait_err = identity_waiter.Wait(ctx)

		// wait success. identity received.
		if wait_err == nil {
			return identity, nil
		}

		// wait failed. considering retry.
		switch we := wait_err.(type) {
		case *waiter.WaiterError:
			// waiter closed (rare occasion). try again.
			continue
		default:
			// context timeout.
			return identity, we
		}
	}
}

// AskDialingPermissionAndGetIdentity returns (nil, false) if the dialing is considered redundant,
// or the peer id is unknown.
// As there is no occasion where a node binds to multiple ports in same host,
// we only compare IP addresses.
func (m *DialInfoMap) AskDialingPermissionAndGetIdentity(id string, addr netip.Addr) (*sec.AbyssPeerIdentity, *DialError) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// Cannot dial if the peer is unknown.
	identity, ok := m.known[id]
	if !ok {
		return nil, &DialError{T: DE_UnknownPeer}
	}

	// There is no need to dial the same IP address twice.
	dialed, ok := m.dialed[id]
	if ok {
		for _, v := range dialed {
			if v.Compare(addr) != 0 {
				return nil, &DialError{T: DE_Redundant}
			}
		}
	}

	return identity, nil
}

// ReportDialTermination removes entry from m.dialed map, allowing retry.
func (m *DialInfoMap) ReportDialTermination(id string, addr netip.Addr) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	dialed, ok := m.dialed[id]
	if !ok {
		return
	}
	for i, v := range dialed {
		if v.Compare(addr) != 0 {
			dialed = slices.Delete(dialed, i, i+1)
		}
	}
}
