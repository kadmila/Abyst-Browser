package ahost_test

import (
	"context"
	"testing"
	"time"

	"github.com/kadmila/Abyss-Browser/abyss_core/ahost"
	"github.com/kadmila/Abyss-Browser/abyss_core/sec"
)

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

	// This should raise EPeerConnected event in one second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	peer_connected := false
	for !peer_connected {
		select {
		case <-ctx.Done():
			t.Fatal("timeout: EPeerConnected event not received within 1 second")
		case event := <-host_A.GetEventCh():
			if e, ok := event.(*ahost.EPeerConnected); ok {
				if e.Peer.ID() == host_B.ID() {
					peer_connected = true
				}
			}
		}
	}
}
