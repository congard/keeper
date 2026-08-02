package http

import "keeper/kit/sender"

type Sender struct {
	endpoint string
	client   *Client
}

func NewSender(endpoint string, client *Client) *Sender {
	return &Sender{
		endpoint: endpoint,
		client:   client,
	}
}

func (s *Sender) Send(payload any) (sender.Status, error) {
	return s.client.Send(s.endpoint, payload)
}
