package poller

import (
	"context"
	"time"

	"keeper/pkg/eventbus"
)

type Worker[T any] interface {
	Do(ctx context.Context) (T, error)
}

type Poller[T any] struct {
	workers  []Worker[T]
	eventBus *eventbus.EventBus[[]T]
	interval time.Duration
}

func New[T any](interval time.Duration) *Poller[T] {
	return &Poller[T]{
		workers:  make([]Worker[T], 0),
		eventBus: eventbus.New[[]T](),
		interval: interval,
	}
}

func (p *Poller[T]) AddWorker(w Worker[T]) {
	p.workers = append(p.workers, w)
}

func (p *Poller[T]) EventBroadcaster() eventbus.Broadcaster[[]T] {
	return p.eventBus
}

func (p *Poller[T]) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			p.tick(ctx)
		}
	}
}

func (p *Poller[T]) tick(ctx context.Context) {
	batch := make([]T, 0, len(p.workers))

	for _, w := range p.workers {
		item, err := w.Do(ctx)
		if err != nil {
			continue
		}
		batch = append(batch, item)
	}

	p.eventBus.Publish(batch)
}
