package transport

type Client interface {
	Send(path string, payload any) (Status, error)
}
