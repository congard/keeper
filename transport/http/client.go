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

type ResponseDecoder = func(io.Reader) (any, error)

func NewClient(url, senderID string) *Client {
	return &Client{
		url:      url,
		senderID: senderID,
		client:   &http.Client{Timeout: 10 * time.Second},
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
