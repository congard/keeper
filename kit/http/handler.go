package http

import (
	"encoding/json"
	"keeper/kit/eventbus"
	"keeper/kit/kitlog"
	"net/http"
)

type Handler interface {
	Endpoint() string
	Handle(http.ResponseWriter, *http.Request)
}

type BroadcastingHandler[T any] struct {
	endpoint string
	eventBus *eventbus.EventBus[Message[T]]
}

func NewBroadcastingHandler[T any](endpoint string) *BroadcastingHandler[T] {
	return &BroadcastingHandler[T]{
		endpoint: endpoint,
		eventBus: eventbus.New[Message[T]](),
	}
}

func (h *BroadcastingHandler[T]) Endpoint() string {
	return h.endpoint
}

func (h *BroadcastingHandler[T]) Handle(w http.ResponseWriter, r *http.Request) {
	var msg Message[T]
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		kitlog.Error(err, kitlog.WithDescription("failed to decode request body"))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.eventBus.Publish(msg)
}

func (h *BroadcastingHandler[T]) Broadcaster() eventbus.Broadcaster[Message[T]] {
	return h.eventBus
}
