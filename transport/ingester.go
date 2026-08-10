package transport

import "keeper/pkg/trie"

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

func (ingester *Ingester) Push(request Request, response Response) {
	exchange, accept := ingester.prePush(Exchange{request, response})
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
