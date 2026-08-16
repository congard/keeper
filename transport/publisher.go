package transport

import (
	"context"
	"log/slog"
	"reflect"

	"keeper/pkg/eventbus"
)

type Publisher[T any] struct {
	sender      Sender[T, Status]
	broadcaster eventbus.Broadcaster[T]
}

func NewPublisher[T any](sender Sender[T, Status], broadcaster eventbus.Broadcaster[T]) *Publisher[T] {
	return &Publisher[T]{
		sender:      sender,
		broadcaster: broadcaster,
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
			response, err := p.sender.Send(value)
			switch {
			case err != nil:
				log.Error("error sending", "error", err)
			case response.IsEmpty():
				log.Error("empty response")
			default:
				status := response.Payload
				log.Info("status",
					slog.String("type", status.Type.String()),
					slog.String("message", status.Message),
					slog.Time("timestamp", response.Timestamp),
				)
			}
		}
	}
}
