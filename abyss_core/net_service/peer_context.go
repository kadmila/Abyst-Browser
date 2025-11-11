package net_service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const INACTIVE_TIMEOUT_MIMUTE = time.Minute * 5
const ACTIVE_TIMEOUT_MINUTE = time.Minute * 180

type Activity struct {
	active_cnt   int
	LastActivity time.Time
}

type ContextedPeer struct {
	ctx        context.Context
	cancelfunc func()
	activity   atomic.Value //Activity
	*AbyssPeer
}

func (c *ContextedPeer) Activate() {
	for {
		prev := c.activity.Load().(Activity)

		next := Activity{
			active_cnt:   prev.active_cnt + 1,
			LastActivity: time.Now(),
		}

		if c.activity.CompareAndSwap(prev, next) {
			return
		}
	}
}

func (c *ContextedPeer) Renew() {
	for {
		prev := c.activity.Load().(Activity)

		next := Activity{
			active_cnt:   prev.active_cnt,
			LastActivity: time.Now(),
		}

		if c.activity.CompareAndSwap(prev, next) {
			return
		}
	}
}

func (c *ContextedPeer) Deactivate() {
	for {
		prev := c.activity.Load().(Activity)

		next := Activity{
			active_cnt:   prev.active_cnt - 1,
			LastActivity: time.Now(),
		}

		if next.active_cnt < 0 {
			panic("invalid behavior:: you deactivated a peer context more than you activated it")
		}

		if c.activity.CompareAndSwap(prev, next) {
			return
		}
	}
}

func (c *ContextedPeer) Context() context.Context {
	return c.ctx
}

func (c *ContextedPeer) Error() error {
	return c.err
}

type ContextedPeerWaiterInfo struct {
	ctx context.Context
	ch  chan *ContextedPeer
}

type ContextedPeerMap struct {
	peers   map[string]*ContextedPeer
	waiters map[string][]*ContextedPeerWaiterInfo

	mtx *sync.Mutex
}

func NewContextedPeerMap() *ContextedPeerMap {
	result := &ContextedPeerMap{
		peers:   make(map[string]*ContextedPeer),
		waiters: make(map[string][]*ContextedPeerWaiterInfo),
		mtx:     new(sync.Mutex),
	}

	return result
}

func (m *ContextedPeerMap) Cleaner(ctx context.Context) {
	for {
		select {
		case <-time.After(INACTIVE_TIMEOUT_MIMUTE):
			m.mtx.Lock()

			t_now := time.Now()

			for _, p := range m.peers {
				if p.ctx.Err() != nil {
					continue
				}

				activity := p.activity.Load().(Activity)
				age := t_now.Sub(activity.LastActivity)
				if age > ACTIVE_TIMEOUT_MINUTE ||
					(activity.active_cnt == 0 && age > INACTIVE_TIMEOUT_MIMUTE) {

					p.cancelfunc()
				}
			}

			dead_wait_targets := make([]string, 0, len(m.waiters))
			for id, wait_group := range m.waiters {
				live_waiters := make([]*ContextedPeerWaiterInfo, 0, len(wait_group))
				for _, waiter := range wait_group {
					if waiter.ctx.Err() == nil {
						live_waiters = append(live_waiters, waiter)
					}
				}
				if len(live_waiters) == 0 {
					dead_wait_targets = append(dead_wait_targets, id)
				} else {
					m.waiters[id] = live_waiters
				}
			}

			for _, dwid := range dead_wait_targets {
				delete(m.waiters, dwid)
			}

			m.mtx.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (m *ContextedPeerMap) Append(ctx context.Context, id string, peer *AbyssPeer) (*ContextedPeer, bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.peers[id]; ok {
		return nil, false
	}

	ctx_new, cf := context.WithCancel(ctx)
	result := &ContextedPeer{
		ctx:        ctx_new,
		cancelfunc: cf,
		AbyssPeer:  peer,
	}
	result.activity.Store(Activity{
		active_cnt:   0,
		LastActivity: time.Now(),
	})
	m.peers[id] = result
	if waiters, ok := m.waiters[id]; ok {
		for _, waiter := range waiters {
			waiter.ch <- result
		}
		delete(m.waiters, id)
	}

	return result, true
}

func (m *ContextedPeerMap) Find(id string) (*ContextedPeer, bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	info, ok := m.peers[id]
	if ok {
		info.Renew()
	}
	return info, ok
}

func (m *ContextedPeerMap) Wait(ctx context.Context, id string) (*ContextedPeer, error) {

	//***Caution***
	m.mtx.Lock()
	info, ok := m.peers[id]
	if ok {
		m.mtx.Unlock()
		return info, nil
	}

	waiters, ok := m.waiters[id]
	if !ok {
		waiters = make([]*ContextedPeerWaiterInfo, 0, 1)
	}
	waiter := &ContextedPeerWaiterInfo{
		ctx: ctx,
		ch:  make(chan *ContextedPeer, 1),
	}
	waiters = append(waiters, waiter)
	m.waiters[id] = waiters
	m.mtx.Unlock()
	//***Caution***
	//the channel is always created, and not discared here.

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-waiter.ch:
		return p, nil
	}
}

///////////// timeout group

type ContextPack struct {
	ctx       context.Context
	cancel    func()
	timestamp time.Time
}

type TimeoutContextGroup struct {
	elem []*ContextPack
	mtx  *sync.Mutex
}

func NewTimeoutContextGroup() *TimeoutContextGroup {
	return &TimeoutContextGroup{
		elem: make([]*ContextPack, 0),
		mtx:  new(sync.Mutex),
	}
}

func (g *TimeoutContextGroup) Cleaner(ctx context.Context, timeout time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(timeout):
			g.mtx.Lock()

			t_now := time.Now()

			live_elem := make([]*ContextPack, 0)
			for _, e := range g.elem {
				if e.ctx.Err() != nil {
					//do nothing
				} else if e.timestamp.Sub(t_now) > timeout {
					e.cancel()
				} else {
					live_elem = append(live_elem, e)
				}
			}

			g.elem = live_elem

			g.mtx.Unlock()
		}
	}
}

func (g *TimeoutContextGroup) Instanciate(background_ctx context.Context) (context.Context, func()) {
	g.mtx.Lock()
	defer g.mtx.Unlock()

	ctx, cancel := context.WithCancel(background_ctx)
	g.elem = append(g.elem, &ContextPack{
		ctx:       ctx,
		cancel:    cancel,
		timestamp: time.Now(),
	})
	return ctx, cancel
}
