package transport

type Sender interface {
	Send(payload any) (Status, error)
}

type ClientSender struct {
	path   string
	client Client
}

func NewClientSender(path string, client Client) *ClientSender {
	return &ClientSender{
		path:   path,
		client: client,
	}
}

func (s *ClientSender) Send(payload any) (Status, error) {
	return s.client.Send(s.path, payload)
}
