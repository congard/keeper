package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestPublishWaitingSingleSubscriber(t *testing.T) {
	eb := New[string]()
	sub := eb.Subscribe()

	done := make(chan string, 1)
	go func() {
		done <- <-sub.Chan()
	}()

	eb.PublishWaiting("hello")

	select {
	case msg := <-done:
		if msg != "hello" {
			t.Fatalf("expected 'hello', got %q", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message on channel")
	}
}

func TestPublishWaitingMultipleSubscribers(t *testing.T) {
	eb := New[int]()
	sub1 := eb.Subscribe()
	sub2 := eb.Subscribe()
	sub3 := eb.Subscribe()

	var received [3]int
	var wg sync.WaitGroup
	channels := []<-chan int{sub1.Chan(), sub2.Chan(), sub3.Chan()}

	for i, ch := range channels {
		wg.Go(func() {
			received[i] = <-ch
		})
	}

	eb.PublishWaiting(42)

	wg.Wait()

	for i, v := range received {
		if v != 42 {
			t.Fatalf("subscriber %d: expected 42, got %d", i, v)
		}
	}
}

func TestUnsubscribeClosesChannel(t *testing.T) {
	eb := New[string]()
	sub := eb.Subscribe()

	received := make(chan string)
	done := make(chan struct{})

	go func() {
		for r := range sub.Chan() {
			received <- r
		}
		done <- struct{}{}
	}()

	eb.PublishWaiting("message")

	msg := <-received
	if msg != "message" {
		t.Fatalf("expected 'message', got %q", msg)
	}

	eb.Unsubscribe(sub)

	select {
	case <-done:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for finishing")
	}
}

func TestUnsubscribeNonExistentSubscriber(t *testing.T) {
	eb := New[string]()
	sub := eb.Subscribe()

	if !eb.Unsubscribe(sub) {
		t.Fatal("expected Unsubscribe to return true for existing subscriber")
	}

	if eb.Unsubscribe(sub) {
		t.Fatal("expected Unsubscribe to return false for already-removed subscriber")
	}
}
