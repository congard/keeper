package direct

import (
	"fmt"
	"keeper/transport"
	"reflect"
)

type Sender[P any, R any] struct {
	senderID string
	route    transport.Route
	ingester *transport.Ingester
}

type request struct {
	route   transport.Route
	payload any
}

type responseData[R any] struct {
	message transport.Message[R]
	err     error
}

type response[R any] struct {
	ch       chan<- responseData[R]
	senderID string
}

func NewSender[P any, R any](senderID string, route transport.Route, ingester *transport.Ingester) *Sender[P, R] {
	return &Sender[P, R]{
		senderID: senderID,
		route:    route,
		ingester: ingester,
	}
}

func (s *Sender[P, R]) Send(payload P) (transport.Message[R], error) {
	ch := make(chan responseData[R], 1)

	func() {
		defer close(ch)

		req := &request{
			route:   s.route,
			payload: transport.NewMessage(s.senderID, payload),
		}
		resp := &response[R]{
			ch:       ch,
			senderID: s.senderID,
		}

		s.ingester.Push(transport.Exchange{
			Request:  req,
			Response: resp,
		})
	}()

	r, ok := <-ch
	if !ok {
		// empty response
		var zero transport.Message[R]
		return zero, nil
	}
	if r.err != nil {
		var zero transport.Message[R]
		return zero, r.err
	}
	return r.message, nil
}

func (r *request) Route() transport.Route {
	return r.route
}

func (r *request) Payload() (any, error) {
	return r.payload, nil
}

func (r *response[R]) Write(payload any) error {
	resp, ok := payload.(R)
	if !ok {
		err := &transport.UnexpectedResponseTypeError{
			Expected: reflect.TypeFor[R]().String(),
			Actual:   fmt.Sprintf("%T", payload),
		}
		r.ch <- responseData[R]{err: err}
		return err
	}
	r.ch <- responseData[R]{message: transport.NewMessage(r.senderID, resp)}
	return nil
}
