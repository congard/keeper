package transport

import (
	"encoding/json"
	"time"
)

type Message[T any] struct {
	SenderID  string
	Timestamp time.Time
	Payload   T
}

type AnyMessage Message[any]
type JSONMessage Message[json.RawMessage]

func NewMessage[T any](senderID string, payload T) Message[T] {
	return Message[T]{
		SenderID:  senderID,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

func (message Message[T]) IsEmpty() bool {
	return message.SenderID == ""
}

func (message Message[T]) ToAny() AnyMessage {
	return AnyMessage{
		SenderID:  message.SenderID,
		Timestamp: message.Timestamp,
		Payload:   message.Payload,
	}
}
