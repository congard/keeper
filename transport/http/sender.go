package http

import (
	"fmt"
	"keeper/transport"
	"reflect"
)

type Sender[P any, R any] struct {
	path    string
	client  *Client
	decoder ResponseDecoder
}

func NewSender[P any, R any](path string, client *Client) *Sender[P, R] {
	return &Sender[P, R]{
		path:    path,
		client:  client,
		decoder: NewResponseDecoder[R](),
	}
}

func (s *Sender[P, R]) Send(payload P) (transport.Message[R], error) {
	result, err := s.client.Send(s.path, payload, s.decoder)
	if err != nil {
		var zero transport.Message[R]
		return zero, err
	}
	// Empty response (nil, nil)
	if result == nil {
		var zero transport.Message[R]
		return zero, nil
	}
	resp, ok := result.(transport.Message[R])
	if !ok {
		var zero transport.Message[R]
		return zero, &transport.UnexpectedResponseTypeError{
			Expected: reflect.TypeFor[transport.Message[R]]().String(),
			Actual:   fmt.Sprintf("%T", result),
		}
	}
	return resp, nil
}
