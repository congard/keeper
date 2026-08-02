package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"keeper/kit/sender"
	"net/http"
	"time"
)

type Client struct {
	url      string
	senderID string
	client   *http.Client
}

func NewClient(url, senderID string) *Client {
	return &Client{
		url:      url,
		senderID: senderID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func handleResponse(r io.Reader) (sender.Status, error) {
	var msg sender.Status
	if err := json.NewDecoder(r).Decode(&msg); err != nil {
		return sender.Status{}, err
	}
	return msg, nil
}

func (c *Client) Send(endpoint string, payload any) (sender.Status, error) {
	body, err := json.Marshal(NewMessage(c.senderID, payload))
	if err != nil {
		return sender.Status{}, fmt.Errorf("marshal message: %w", err)
	}

	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.url, endpoint), bodyReader)
	if err != nil {
		return sender.Status{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return sender.Status{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return sender.Status{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return handleResponse(resp.Body)
}
