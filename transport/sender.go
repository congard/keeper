package transport

type Sender[P any, R any] interface {
	Send(payload P) (Message[R], error)
}
