package transport

import (
	"context"
	"log/slog"
	"reflect"

	"keeper/pkg/eventbus"
	"keeper/pkg/logger"
)

type Publisher[T any] struct {
	sender      Sender
	broadcaster eventbus.Broadcaster[T]
	transform   transformer[T]
}

type PublisherOption[T any] func(*publisherOptions[T])

type publisherOptions[T any] struct {
	transform transformer[T]
}

type transformer[T any] func(T) any

func NewPublisher[T any](
	sender Sender,
	broadcaster eventbus.Broadcaster[T],
	opts ...PublisherOption[T],
) *Publisher[T] {
	po := &publisherOptions[T]{
		transform: func(t T) any { return t },
	}
	for _, opt := range opts {
		opt(po)
	}

	return &Publisher[T]{
		sender:      sender,
		broadcaster: broadcaster,
		transform:   po.transform,
	}
}

func WithTransformer[T any](t transformer[T]) PublisherOption[T] {
	return func(po *publisherOptions[T]) {
		po.transform = t
	}
}

func (p *Publisher[T]) Run(ctx context.Context) error {
	sub := p.broadcaster.Subscribe()
	defer sub.Unsubscribe()

	log := slog.With("component", "publisher", "endpoint", reflect.TypeOf(p.sender).String())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case value := <-sub.Chan():
			status, err := p.sender.Send(p.transform(value))
			if err != nil {
				logger.LogIfError(err)
			} else {
				log.Info("status",
					slog.String("type", status.Type.String()),
					slog.String("message", status.Message),
					slog.Time("timestamp", status.Timestamp),
				)
			}
		}
	}
}
