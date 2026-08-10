package http

import (
	"encoding/json"
	"keeper/kit/eventbus"
	"keeper/kit/kitlog"
	"keeper/transport"
	"net/http"
)

type Handler interface {
	Route() string
	Handle(http.ResponseWriter, *http.Request)
}

type BroadcastingHandler[T any] struct {
	endpoint string
	eventBus *eventbus.EventBus[transport.Message[T]]
}

func NewBroadcastingHandler[T any](endpoint string) *BroadcastingHandler[T] {
	return &BroadcastingHandler[T]{
		endpoint: endpoint,
		eventBus: eventbus.New[transport.Message[T]](),
	}
}

func (h *BroadcastingHandler[T]) Route() string {
	return h.endpoint
}

func (h *BroadcastingHandler[T]) Handle(w http.ResponseWriter, r *http.Request) {
	var msg transport.Message[T]
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		kitlog.Error(err, kitlog.WithDescription("failed to decode request body"))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.eventBus.Publish(msg)
}

func (h *BroadcastingHandler[T]) Broadcaster() eventbus.Broadcaster[transport.Message[T]] {
	return h.eventBus
}
