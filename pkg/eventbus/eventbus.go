package eventbus

import (
	"sync"
)

type Subscriber[T any] interface {
	Chan() <-chan T
	Unsubscribe() bool
}

type Broadcaster[T any] interface {
	Subscribe() Subscriber[T]
	Unsubscribe(Subscriber[T]) bool
}

type Publisher[T any] interface {
	Publish(e T)
	PublishWaiting(e T)
}

type EventBus[T any] struct {
	mu        sync.RWMutex
	listeners []chan T
}

type subscriber[T any] struct {
	channel  chan T
	eventBus *EventBus[T]
}

func New[T any]() *EventBus[T] {
	return &EventBus[T]{
		listeners: make([]chan T, 0),
	}
}

func (eb *EventBus[T]) Subscribe() Subscriber[T] {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan T)
	eb.listeners = append(eb.listeners, ch)
	return &subscriber[T]{ch, eb}
}

func (eb *EventBus[T]) Unsubscribe(sub Subscriber[T]) bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ebSub, ok := sub.(*subscriber[T])

	if !ok {
		return false
	}

	ch := ebSub.channel

	for i, l := range eb.listeners {
		if l == ch {
			eb.listeners = append(eb.listeners[:i], eb.listeners[i+1:]...)
			close(ch)
			return true
		}
	}

	return false
}

func (eb *EventBus[T]) Publish(e T) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, ch := range eb.listeners {
		go func() {
			ch <- e
		}()
	}
}

func (eb *EventBus[T]) PublishWaiting(e T) {
	eb.mu.RLock()

	var wg sync.WaitGroup

	for _, ch := range eb.listeners {
		wg.Go(func() {
			ch <- e
		})
	}

	eb.mu.RUnlock()

	wg.Wait()
}

func (sub *subscriber[T]) Chan() <-chan T {
	return sub.channel
}

func (sub *subscriber[T]) Unsubscribe() bool {
	return sub.eventBus.Unsubscribe(sub)
}
