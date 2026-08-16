package direct

import (
	"testing"

	"keeper/transport"
	"keeper/transport/sendertest"
)

func newDirectSenderFactory(t *testing.T) sendertest.StringSenderFactory {
	t.Helper()

	return func(_ *testing.T, handler transport.TypedHandlerFunc[string]) transport.Sender[string, string] {
		ingester := transport.NewIngester(nil)
		ingester.Handle(transport.ParseRoute("test"), transport.NewTypedHandler(handler))
		return NewSender[string, string](sendertest.SenderID, transport.ParseRoute("test"), ingester)
	}
}

func TestDirectSenderContract(t *testing.T) {
	factory := newDirectSenderFactory(t)
	sendertest.RunSenderContractTests(t, factory)
}
