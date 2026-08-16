package transport

import (
	"fmt"
	"keeper/pkg/trie"
	"reflect"
)

type Request interface {
	Route() Route
	Payload() (any, error)
}

type Response interface {
	Write(payload any) error
}

type Exchange struct {
	Request  Request
	Response Response
}

type TypedRequest[T any] struct {
	request Request
}

type TypedHandlerFunc[T any] func(TypedRequest[T], Response)

type HandlerFunc func(Request, Response)
type PrePushFunc func(Exchange) (exchange Exchange, accept bool)

type Ingester struct {
	handlers *trie.Trie[RouteSegment, HandlerFunc]
	prePush  PrePushFunc
}

type IngesterConfig struct {
	PrePush PrePushFunc
}

func NewIngester(config *IngesterConfig) *Ingester {
	if config == nil {
		config = &IngesterConfig{}
	}
	if config.PrePush == nil {
		config.PrePush = func(e Exchange) (exchange Exchange, accept bool) {
			exchange = e
			accept = true
			return
		}
	}

	return &Ingester{
		handlers: trie.New[RouteSegment, HandlerFunc](),
		prePush:  config.PrePush,
	}
}

func (ingester *Ingester) Push(exchange Exchange) {
	exchange, accept := ingester.prePush(exchange)
	if !accept {
		return
	}

	handler, ok := ingester.handlers.Value(exchange.Request.Route()...)
	if !ok {
		return
	}

	handler(exchange.Request, exchange.Response)
}

func (ingester *Ingester) Handle(route Route, handler HandlerFunc) {
	ingester.handlers.SetValue(handler, route...)
}

func (r TypedRequest[T]) Route() Route {
	return r.request.Route()
}

func (r TypedRequest[T]) Payload() (Message[T], error) {
	raw, err := r.request.Payload()
	if err != nil {
		var zero Message[T]
		return zero, err
	}
	payload, ok := raw.(Message[T])
	if !ok {
		var zero Message[T]
		return zero, &UnexpectedResponseTypeError{
			Expected: reflect.TypeFor[Message[T]]().String(),
			Actual:   fmt.Sprintf("%T", raw),
		}
	}
	return payload, nil
}

func NewTypedHandler[T any](handler TypedHandlerFunc[T]) HandlerFunc {
	return func(req Request, resp Response) {
		handler(TypedRequest[T]{req}, resp)
	}
}
