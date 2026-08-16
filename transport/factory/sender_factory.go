package factory

import (
	"fmt"
	"keeper/transport"
	"keeper/transport/direct"
	"keeper/transport/http"
	"net/url"
)

const (
	directScheme = "direct"
	httpScheme   = "http"
	httpsScheme  = "https"
)

type SenderConfig struct {
	Ingester   *transport.Ingester
	ClientPool *http.ClientPool
}

func NewSender[P any, R any](senderID, rawURL string, config SenderConfig) (transport.Sender[P, R], error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if len(u.Query()) != 0 || u.Fragment != "" {
		return nil, fmt.Errorf("URL must not contain query parameters or fragment")
	}

	switch u.Scheme {
	case directScheme:
		return direct.NewSender[P, R](senderID, transport.ParseRoute(u.Path), config.Ingester), nil
	case httpScheme, httpsScheme:
		return http.NewSender[P, R](u.Path, config.ClientPool.Acquire(senderID, fmt.Sprintf("%s://%s", u.Scheme, u.Host))), nil
	}

	return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
}
