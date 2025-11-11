package main

import (
	"context"
	"testing"
	"time"
)

func TestSteppingConsumer(t *testing.T) {
	consumer := MakeSteppingConsumer()
	fatal_ch := make(chan string)

	go func() {
		for range 4 {
			{
				_, ok := consumer.TryPut("cat")
				if ok {
					fatal_ch <- "consumer accepted when not consuming"
				}
			}

			done_ch := make(chan bool, 2)
			start_time := time.Now()

			go func() {
				ctx, ctx_cancel := context.WithTimeout(context.Background(), time.Duration(1<<63-1))
				defer ctx_cancel()

				arg, consume, ok, ok_mono := consumer.TryConsume(ctx)
				if !ok_mono {
					fatal_ch <- "duplicate wait"
				}

				if !ok {
					fatal_ch <- "unexpected timeout"
				}

				get_duration := time.Since(start_time)
				if arg != "cat" {
					fatal_ch <- "data corruption"
				}

				if (get_duration < time.Millisecond*100) || (get_duration > time.Millisecond*150) {
					fatal_ch <- "consume duration mismatch"
				}

				<-time.After(time.Millisecond * 500)
				consume <- true

				done_ch <- true
			}()

			go func() {
				<-time.After(time.Millisecond * 100)

				consumed, ok := consumer.TryPut("cat")
				if !ok {
					fatal_ch <- "failed to put"
				}

				_, ok = consumer.TryPut("another")
				if ok {
					fatal_ch <- "duplicate put accepted"
				}

				select {
				case consume_res := <-consumed:
					consume_duration := time.Since(start_time)
					if (consume_duration < time.Millisecond*600) || (consume_duration > time.Millisecond*700) {
						fatal_ch <- "consume duration (sender) mismatch"
					}
					if !consume_res {
						fatal_ch <- "consume result false"
					}
				case <-time.After(time.Millisecond * 1000):
					fatal_ch <- "consume chan still blocked after consuming"
				}

				_, ok = consumer.TryPut("third")
				if ok {
					fatal_ch <- "(third) duplicate put accepted"
				}

				done_ch <- true
			}()

			<-done_ch
			<-done_ch
		}
		close(fatal_ch)
	}()

	fatal, ok := <-fatal_ch
	if ok {
		t.Fatal(fatal)
	}
}
