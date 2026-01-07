package ahost_test

import (
	"context"
	"testing"
	"time"

	"github.com/kadmila/Abyss-Browser/abyss_core/ahost"
	"github.com/kadmila/Abyss-Browser/abyss_core/and"
	"github.com/kadmila/Abyss-Browser/abyss_core/sec"
)

// expectEvent waits for an event of type T.
// Returns the event on success or fails the test on timeout.
// Skips events that don't match the expected type.
func expectEvent[T any](t *testing.T, event_ch <-chan any) T {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var zero T
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for event %T", zero)
			return zero
		case event := <-event_ch:
			if typed_event, ok := event.(T); ok {
				return typed_event
			}
			// Skip events that don't match - they might be for other purposes
		}
	}
}

func TestPeerConnectedEvent(t *testing.T) {
	// Construct two hosts
	root_key_A, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	host_A, err := ahost.NewAbyssHost(root_key_A)
	if err != nil {
		t.Fatal(err)
	}

	root_key_B, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	host_B, err := ahost.NewAbyssHost(root_key_B)
	if err != nil {
		t.Fatal(err)
	}

	// Bind both hosts
	err = host_A.Bind()
	if err != nil {
		t.Fatal(err)
	}
	err = host_B.Bind()
	if err != nil {
		t.Fatal(err)
	}

	// Start serving (Serve() blocks, so run in goroutines)
	go host_A.Serve()
	go host_B.Serve()
	defer host_A.Close()
	defer host_B.Close()

	// Exchange peer information
	err = host_A.AppendKnownPeer(host_B.RootCertificate(), host_B.HandshakeKeyCertificate())
	if err != nil {
		t.Fatal(err)
	}
	err = host_B.AppendKnownPeer(host_A.RootCertificate(), host_A.HandshakeKeyCertificate())
	if err != nil {
		t.Fatal(err)
	}

	// One host dials another
	err = host_A.Dial(host_B.ID())
	if err != nil {
		t.Fatal(err)
	}

	// This should raise EPeerConnected event
	expectEvent[*ahost.EPeerConnected](t, host_A.GetEventCh())
	expectEvent[*ahost.EPeerConnected](t, host_B.GetEventCh())
}

func TestJoinWorld(t *testing.T) {
	// Construct two hosts
	root_key_A, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	host_A, err := ahost.NewAbyssHost(root_key_A)
	if err != nil {
		t.Fatal(err)
	}

	root_key_B, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	host_B, err := ahost.NewAbyssHost(root_key_B)
	if err != nil {
		t.Fatal(err)
	}

	// Bind both hosts
	err = host_A.Bind()
	if err != nil {
		t.Fatal(err)
	}
	err = host_B.Bind()
	if err != nil {
		t.Fatal(err)
	}

	// Start serving
	go host_A.Serve()
	go host_B.Serve()
	defer host_A.Close()
	defer host_B.Close()

	// Exchange peer information
	err = host_A.AppendKnownPeer(host_B.RootCertificate(), host_B.HandshakeKeyCertificate())
	if err != nil {
		t.Fatal(err)
	}
	err = host_B.AppendKnownPeer(host_A.RootCertificate(), host_A.HandshakeKeyCertificate())
	if err != nil {
		t.Fatal(err)
	}

	// Host A opens a world
	world_A := host_A.OpenWorld("abyss://example.com/test")
	if world_A == nil {
		t.Fatal("world should not be nil")
	}

	// Expose world to default path "/"
	host_A.ExposeWorldForJoin(world_A, "/")

	// Wait for world enter event on host A
	expectEvent[*and.EANDWorldEnter](t, host_A.GetEventCh())

	// Host B dials host A
	err = host_B.Dial(host_A.ID())
	if err != nil {
		t.Fatal(err)
	}

	// Wait for peer connection on host B
	peer_B_to_A := expectEvent[*ahost.EPeerConnected](t, host_B.GetEventCh())
	expectEvent[*ahost.EPeerConnected](t, host_A.GetEventCh())

	// Host B joins the world at path "/"
	host_B.JoinWorld(peer_B_to_A.Peer, "/")

	// Wait for session request event on host A (host A should receive join request)
	session_request := expectEvent[*and.EANDSessionRequest](t, host_A.GetEventCh())

	// Host A accepts the session request
	host_A.AcceptWorldSession(world_A, session_request.Peer, session_request.SessionID)

	// Host B should receive EANDWorldEnter event
	expectEvent[*and.EANDWorldEnter](t, host_B.GetEventCh())

	// Both hosts should receive EANDSessionReady event
	expectEvent[*and.EANDSessionReady](t, host_A.GetEventCh())
	expectEvent[*and.EANDSessionReady](t, host_B.GetEventCh())
}

func TestJoinWorldTransitive(t *testing.T) {
	// Construct three hosts
	root_key_A, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	host_A, err := ahost.NewAbyssHost(root_key_A)
	if err != nil {
		t.Fatal(err)
	}

	root_key_B, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	host_B, err := ahost.NewAbyssHost(root_key_B)
	if err != nil {
		t.Fatal(err)
	}

	root_key_C, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	host_C, err := ahost.NewAbyssHost(root_key_C)
	if err != nil {
		t.Fatal(err)
	}

	// Bind all hosts
	if err = host_A.Bind(); err != nil {
		t.Fatal(err)
	}
	if err = host_B.Bind(); err != nil {
		t.Fatal(err)
	}
	if err = host_C.Bind(); err != nil {
		t.Fatal(err)
	}

	// Start serving
	go host_A.Serve()
	go host_B.Serve()
	go host_C.Serve()
	defer host_A.Close()
	defer host_B.Close()
	defer host_C.Close()

	// Exchange peer information (A-B and B-C, but not A-C initially)
	if err = host_A.AppendKnownPeer(host_B.RootCertificate(), host_B.HandshakeKeyCertificate()); err != nil {
		t.Fatal(err)
	}
	if err = host_B.AppendKnownPeer(host_A.RootCertificate(), host_A.HandshakeKeyCertificate()); err != nil {
		t.Fatal(err)
	}
	if err = host_B.AppendKnownPeer(host_C.RootCertificate(), host_C.HandshakeKeyCertificate()); err != nil {
		t.Fatal(err)
	}
	if err = host_C.AppendKnownPeer(host_B.RootCertificate(), host_B.HandshakeKeyCertificate()); err != nil {
		t.Fatal(err)
	}

	// Synchronization channels
	world_exposed_A := make(chan struct{})
	world_exposed_B := make(chan struct{})
	B_joined := make(chan struct{})

	// Run each host's event handling in separate goroutines
	done_A := make(chan error, 1)
	done_B := make(chan error, 1)
	done_C := make(chan error, 1)

	// Host A goroutine
	go func() {
		defer func() { done_A <- nil }()

		// 1. Opens world, exposes at "/" -> EANDWorldEnter
		world_A := host_A.OpenWorld("abyss://example.com/transitive")
		host_A.ExposeWorldForJoin(world_A, "/")
		expectEvent[*and.EANDWorldEnter](t, host_A.GetEventCh())
		close(world_exposed_A)

		// 2. EPeerConnected (B)
		expectEvent[*ahost.EPeerConnected](t, host_A.GetEventCh())

		// 3. receives EANDSessionRequest (from B), accepts it
		session_req_B := expectEvent[*and.EANDSessionRequest](t, host_A.GetEventCh())
		host_A.AcceptWorldSession(world_A, session_req_B.Peer, session_req_B.SessionID)

		// 4. receives EANDSessionReady (from B)
		expectEvent[*and.EANDSessionReady](t, host_A.GetEventCh())

		// 5. EPeerConnected (C)
		expectEvent[*ahost.EPeerConnected](t, host_A.GetEventCh())

		// 6. receives EANDSessionRequest (from C), accepts it
		session_req_C := expectEvent[*and.EANDSessionRequest](t, host_A.GetEventCh())
		host_A.AcceptWorldSession(world_A, session_req_C.Peer, session_req_C.SessionID)

		// 7. receives EANDSessionReady (from C)
		expectEvent[*and.EANDSessionReady](t, host_A.GetEventCh())
	}()

	// Host B goroutine
	go func() {
		defer func() { done_B <- nil }()

		// Wait for A to expose the world
		<-world_exposed_A

		// 1. dials A -> EPeerConnected (A)
		if err := host_B.Dial(host_A.ID()); err != nil {
			t.Error(err)
			return
		}
		peer_B_to_A := expectEvent[*ahost.EPeerConnected](t, host_B.GetEventCh())

		// 2. joins "/" to A -> EANDWorldEnter -> exposes at "/shared"
		host_B.JoinWorld(peer_B_to_A.Peer, "/")
		world_enter_B := expectEvent[*and.EANDWorldEnter](t, host_B.GetEventCh())
		host_B.ExposeWorldForJoin(world_enter_B.World, "/shared")
		close(world_exposed_B)

		// 3. EANDSessionReady (from A)
		expectEvent[*and.EANDSessionReady](t, host_B.GetEventCh())
		close(B_joined)

		// 4. EPeerConnected (C)
		expectEvent[*ahost.EPeerConnected](t, host_B.GetEventCh())

		// 5. receives EANDSessionRequest (from C), accepts it
		session_req_C := expectEvent[*and.EANDSessionRequest](t, host_B.GetEventCh())
		host_B.AcceptWorldSession(world_enter_B.World, session_req_C.Peer, session_req_C.SessionID)

		// 6. receives EANDSessionReady (from C)
		expectEvent[*and.EANDSessionReady](t, host_B.GetEventCh())
	}()

	// Host C goroutine
	go func() {
		defer func() { done_C <- nil }()

		// Wait for B to join and expose the world
		<-world_exposed_B
		<-B_joined

		// 1. dials B -> EPeerConnected (B)
		if err := host_C.Dial(host_B.ID()); err != nil {
			t.Error(err)
			return
		}
		peer_C_to_B := expectEvent[*ahost.EPeerConnected](t, host_C.GetEventCh())

		// 2. joins "/shared" to B -> EANDWorldEnter
		host_C.JoinWorld(peer_C_to_B.Peer, "/shared")
		world_enter_C := expectEvent[*and.EANDWorldEnter](t, host_C.GetEventCh())

		// 3. EANDSessionReady (from B)
		expectEvent[*and.EANDSessionReady](t, host_C.GetEventCh())

		// 4. EPeerConnected (A)
		expectEvent[*ahost.EPeerConnected](t, host_C.GetEventCh())

		// 5. receives EANDSessionRequest (from A), accepts it
		session_req_A := expectEvent[*and.EANDSessionRequest](t, host_C.GetEventCh())
		host_C.AcceptWorldSession(world_enter_C.World, session_req_A.Peer, session_req_A.SessionID)

		// 6. receives EANDSessionReady (from A)
		expectEvent[*and.EANDSessionReady](t, host_C.GetEventCh())
	}()

	// Wait for all goroutines to complete
	timeout := time.After(time.Second * 10)
	for range 3 {
		select {
		case err := <-done_A:
			if err != nil {
				t.Fatalf("Host A failed: %v", err)
			}
		case err := <-done_B:
			if err != nil {
				t.Fatalf("Host B failed: %v", err)
			}
		case err := <-done_C:
			if err != nil {
				t.Fatalf("Host C failed: %v", err)
			}
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}
}
