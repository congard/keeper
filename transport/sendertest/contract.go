package sendertest

import (
	"encoding/json"
	"testing"

	"keeper/pkg/errors"
	"keeper/transport"
)

type SenderFactory[P any, R any] func(t *testing.T, handler transport.TypedHandlerFunc[P]) transport.Sender[P, R]
type StringSenderFactory SenderFactory[string, string]

const SenderID = "sender-1"

func RunSenderContractTests(t *testing.T, factory StringSenderFactory) {
	t.Helper()

	t.Run("SendReturnsResponse", func(t *testing.T) {
		testSendReturnsResponse(t, factory)
	})

	t.Run("SendPropagatesSenderID", func(t *testing.T) {
		testSendPropagatesSenderID(t, factory)
	})

	t.Run("SendWrongResponseType", func(t *testing.T) {
		testSendWrongResponseType(t, factory)
	})

	t.Run("SendEmptyResponse", func(t *testing.T) {
		testSendEmptyResponse(t, factory)
	})
}

func testSendReturnsResponse(t *testing.T, factory StringSenderFactory) {
	t.Helper()

	sender := factory(t, func(req transport.TypedRequest[string], resp transport.Response) {
		msg, err := req.Payload()
		if err != nil {
			t.Fatalf("req.Payload() error = %v, want nil", err)
		}
		if msg.Payload != "ping" {
			t.Errorf("request payload = %q, want %q", msg.Payload, "ping")
		}
		_ = resp.Write("pong")
	})

	msg, err := sender.Send("ping")
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	if msg.IsEmpty() {
		t.Errorf("msg.IsEmpty() = true, want false for successful response")
	}
	if msg.Payload != "pong" {
		t.Errorf("msg.Payload = %q, want %q", msg.Payload, "pong")
	}
}

func testSendPropagatesSenderID(t *testing.T, factory StringSenderFactory) {
	t.Helper()

	sender := factory(t, func(req transport.TypedRequest[string], resp transport.Response) {
		msg, err := req.Payload()
		if err != nil {
			t.Fatalf("req.Payload() error = %v, want nil", err)
		}
		if msg.SenderID != SenderID {
			t.Errorf("request SenderID = %q, want %q", msg.SenderID, SenderID)
		}
		if msg.Payload != "ping" {
			t.Errorf("request payload = %q, want %q", msg.Payload, "ping")
		}
		_ = resp.Write("pong")
	})

	_, err := sender.Send("ping")
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
}

func testSendWrongResponseType(t *testing.T, factory StringSenderFactory) {
	t.Helper()

	sender := factory(t, func(_ transport.TypedRequest[string], resp transport.Response) {
		// Write a value that is not string (int instead).
		_ = resp.Write(123)
	})

	_, err := sender.Send("ping")
	if err == nil {
		t.Fatal("Send() error = nil, want non-nil for wrong response type")
	}
	if !errors.IsErrorOfType[*transport.UnexpectedResponseTypeError](err) &&
		!errors.IsErrorOfType[*json.UnmarshalTypeError](err) {
		t.Errorf("Send() error = %v, want an UnexpectedResponseTypeError or json.UnmarshalTypeError", err)
	}
}

func testSendEmptyResponse(t *testing.T, factory StringSenderFactory) {
	t.Helper()

	sender := factory(t, func(_ transport.TypedRequest[string], _ transport.Response) {
		// No Write call - empty response
	})

	msg, err := sender.Send("ping")
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
	if !msg.IsEmpty() {
		t.Errorf("msg.IsEmpty() = false, want true for zero message")
	}
}
