package ann_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kadmila/Abyss-Browser/abyss_core/ani"
	"github.com/kadmila/Abyss-Browser/abyss_core/ann"
	"github.com/kadmila/Abyss-Browser/abyss_core/sec"
)

func TestNewAbyssNode(t *testing.T) {
	// Node construction
	root_key_A, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	var node_A ani.IAbyssNode
	node_A, err = ann.NewAbyssNode(root_key_A)
	if err != nil {
		t.Fatal(err)
	}

	root_key_B, err := sec.NewRootPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	var node_B ani.IAbyssNode
	node_B, err = ann.NewAbyssNode(root_key_B)
	if err != nil {
		t.Fatal(err)
	}

	// Listening
	err = node_A.Listen()
	if err != nil {
		t.Fatal(err)
	}
	err = node_B.Listen()
	if err != nil {
		t.Fatal(err)
	}

	// Start service loop
	node_A_done := make(chan error)
	node_B_done := make(chan error)
	go func() {
		node_A_done <- node_A.Serve()
	}()
	go func() {
		node_B_done <- node_B.Serve()
	}()

	// Appending peer information
	err = node_A.AppendKnownPeer(node_B.RootCertificate(), node_B.HandshakeKeyCertificate())
	if err != nil {
		t.Fatal(err)
	}
	err = node_B.AppendKnownPeer(node_A.RootCertificate(), node_A.HandshakeKeyCertificate())
	if err != nil {
		t.Fatal(err)
	}

	// Mutual dialing (all address candidates)
	for _, v := range node_A.LocalAddrCandidates() {
		fmt.Println(v)
		err = node_B.Dial(node_A.ID(), v)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range node_B.LocalAddrCandidates() {
		fmt.Println(v)
		err = node_A.Dial(node_B.ID(), v)
		if err != nil {
			t.Fatal(err)
		}
	}

	ctx, ctxcancel := context.WithTimeout(context.Background(), time.Second*3)
	defer ctxcancel()

	// Accept for 3 seconds.
	peer_A_B_ch := make(chan ani.IAbyssPeer, 1)
	peer_B_A_ch := make(chan ani.IAbyssPeer, 1)
	go func() {
		for {
			peer_A_B, err := node_A.Accept(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					return
				}
				fmt.Println("accept>> ", err)
			} else {
				fmt.Println("connected: " + peer_A_B.RemoteAddr().String() + peer_A_B.ID())
				peer_A_B_ch <- peer_A_B
			}
		}
	}()
	go func() {
		for {
			peer_B_A, err := node_B.Accept(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					return
				}
				fmt.Println("accept>> ", err)
			} else {
				fmt.Println("connected: " + peer_B_A.RemoteAddr().String() + peer_B_A.ID())
				peer_B_A_ch <- peer_B_A
			}
		}
	}()

	// Concurrently, wait for peers, and check correctness.
	select {
	case <-ctx.Done():
		t.Fatal("accept timeout")
	case peer_A_B := <-peer_A_B_ch:
		if peer_A_B.ID() != node_B.ID() {
			t.Fatal("peer id mismatch")
		}
	}
	select {
	case <-ctx.Done():
		t.Fatal("accept timeout")
	case peer_B_A := <-peer_B_A_ch:
		if peer_B_A.ID() != node_A.ID() {
			t.Fatal("peer id mismatch")
		}
	}

	<-ctx.Done()

	node_A.Close()
	node_B.Close()

	err = <-node_A_done
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	err = <-node_B_done
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
