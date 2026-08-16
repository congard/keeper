package http

import (
	"net/http/httptest"
	"testing"

	"keeper/transport"
	"keeper/transport/sendertest"
)

func newHTTPSenderFactory(t *testing.T) sendertest.StringSenderFactory {
	t.Helper()

	return func(t *testing.T, handler transport.TypedHandlerFunc[string]) transport.Sender[string, string] {
		ingester := transport.NewIngester(nil)
		ingester.Handle(transport.ParseRoute("test"), transport.NewTypedHandler(handler))

		server := NewServer("server-1", ingester)
		server.Handle("/test", NewExchangeAdapter[string](server))

		ts := httptest.NewServer(server.Handler)
		t.Cleanup(func() { ts.Close() })

		client := NewClient(ts.URL, sendertest.SenderID)
		return NewSender[string, string]("test", client)
	}
}

func TestHTTPSenderContract(t *testing.T) {
	factory := newHTTPSenderFactory(t)
	sendertest.RunSenderContractTests(t, factory)
}
