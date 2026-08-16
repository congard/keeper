package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"keeper/transport"
	"net/http"
	"time"
)

type Client struct {
	url      string
	senderID string
	client   *http.Client
}

type ClientConfig struct {
	SenderID string
	URL      string
	Client   *http.Client
}

type ClientPool struct {
	client  *http.Client
	clients []*Client
}

type ResponseDecoder = func(io.Reader) (any, error)

func NewClient(config ClientConfig) *Client {
	client := config.Client
	if client == nil {
		client = defaultHTTPClient()
	}
	return &Client{
		url:      config.URL,
		senderID: config.SenderID,
		client:   client,
	}
}

func (c *Client) Send(path string, payload any, responseDecoder ResponseDecoder) (any, error) {
	body, err := json.Marshal(transport.NewMessage(c.senderID, payload))
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.url, path), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return responseDecoder(resp.Body)
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		client:  defaultHTTPClient(),
		clients: make([]*Client, 0),
	}
}

func (p *ClientPool) Acquire(senderID, url string) *Client {
	for _, c := range p.clients {
		if c.senderID == senderID && c.url == url {
			return c
		}
	}
	c := NewClient(ClientConfig{SenderID: senderID, URL: url, Client: p.client})
	p.clients = append(p.clients, c)
	return c
}

func NewResponseDecoder[T any]() ResponseDecoder {
	return func(reader io.Reader) (any, error) {
		// Peek at the first byte to check if response is empty without consuming the stream
		bufReader := bufio.NewReader(reader)
		firstByte, err := bufReader.Peek(1)
		if err != nil {
			if err == io.EOF {
				return nil, nil // Empty response
			}
			return nil, err
		}
		if len(firstByte) == 0 {
			return nil, nil // Empty response
		}

		var msg transport.Message[T]
		if err := json.NewDecoder(bufReader).Decode(&msg); err != nil {
			return nil, err
		}
		return msg, nil
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}
