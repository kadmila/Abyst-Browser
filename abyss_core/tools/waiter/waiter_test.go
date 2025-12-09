package waiter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kadmila/Abyss-Browser/abyss_core/tools/waiter"
)

func TestWait(t *testing.T) {
	w := waiter.NewWaiter[int]()

	fmt.Println("begin")

	go waitPrint(w)
	go waitPrint(w)
	go waitPrint(w)
	go waitPrint(w)

	<-time.After(time.Second * 3)

	w.Set(17)
	<-time.After(time.Second)
}

func waitPrint(w *waiter.Waiter[int]) {
	v, err := w.Wait(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(v)
}

func TestWait2(t *testing.T) {
	w := waiter.NewWaiter[int]()

	fmt.Println("begin")

	go waitPrint2(w)
	go waitPrint2(w)
	go waitPrint2(w)
	go waitPrint2(w)

	<-time.After(time.Second * 3)

	w.Set(17)
	<-time.After(time.Second)

	waitPrint(w)
}

func waitPrint2(w *waiter.Waiter[int]) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelfunc()

	v, err := w.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(v)
}

func TestTryClose(t *testing.T) {
	w := waiter.NewWaiter[int]()

	go waitPrint(w)

	<-time.After(time.Second)

	wait_cnt, ok := w.TryClose()
	if ok {
		t.Fatal("should not close here")
	}
	fmt.Println(wait_cnt)

	w.Set(32)

	<-time.After(time.Second)

	wait_cnt, ok = w.TryClose()
	if !ok {
		t.Fatal("should close here")
	}
	fmt.Println(wait_cnt)
}
