package main

import (
	"context"
	"sync/atomic"
)

type ConsumeContract struct {
	argument     any
	confirmation chan bool
}

// we expect consumer to always wait on TryConsume()
type SteppingConsumer struct {
	is_open      *atomic.Bool
	ch           chan ConsumeContract
	is_consuming *atomic.Bool
}

func MakeConsumeContract(argument any) ConsumeContract {
	return ConsumeContract{
		argument:     argument,
		confirmation: make(chan bool, 1),
	}
}

func MakeSteppingConsumer() SteppingConsumer {
	res := SteppingConsumer{
		is_open:      &atomic.Bool{},
		ch:           make(chan ConsumeContract, 1),
		is_consuming: &atomic.Bool{},
	}
	return res
}
func (c *SteppingConsumer) TryPut(argument any) (<-chan bool, bool) {
	if !c.is_open.CompareAndSwap(true, false) {
		//failed to occupy. no-op to consumer
		return nil, false
	}

	contract := MakeConsumeContract(argument)
	c.ch <- contract //this must always succeed

	return contract.confirmation, true
}

// calling TryConsume concurrently causes last return value to be false
func (c *SteppingConsumer) TryConsume(ctx context.Context) (any, chan<- bool, bool, bool) {
	if !c.is_consuming.CompareAndSwap(false, true) {
		return nil, nil, false, false
	}
	defer func() {
		c.is_consuming.Store(false)
	}()

	c.is_open.Store(true)

	select {
	case contract := <-c.ch:
		//is_open remains false; blocks additional contracts until TryConsume() re-opens.
		return contract.argument, contract.confirmation, true, true
	case <-ctx.Done():
		// have to close occupation
		if !c.is_open.CompareAndSwap(true, false) {
			//put at last moment. We just handle this as valid.
			contract := <-c.ch
			return contract.argument, contract.confirmation, true, true
		}
		return nil, nil, false, true
	}
}
