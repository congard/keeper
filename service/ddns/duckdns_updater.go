package ddns

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"keeper/pkg/logger"
)

type DuckDNSUpdater struct {
	domain string
	token  string
	client *http.Client
}

func NewDuckDNSUpdater(domain, token, bindIP string) *DuckDNSUpdater {
	return &DuckDNSUpdater{
		domain: domain,
		token:  token,
		client: newClientWithIP(net.ParseIP(bindIP)),
	}
}

func newClientWithIP(ip net.IP) *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if ip != nil {
		dialer.LocalAddr = &net.TCPAddr{
			IP: ip,
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}
}

const (
	updateStatusOk        = "OK"
	updateStatusErr       = "KO"
	updateResultOk        = "UPDATED"
	updateResultUnchanged = "NOCHANGE"
	okResponseLines       = 4
)

func (updater *DuckDNSUpdater) Update() (UpdateResult, error) {
	target := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&verbose=true",
		url.QueryEscape(updater.domain), url.QueryEscape(updater.token))

	resp, err := updater.client.Get(target)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("duckdns request failed: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		logger.LogIfErrorTyped[DuckDNSUpdater](err)
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("duckdns read response failed: %w", err)
	}

	response := strings.TrimSpace(string(body))
	lines := strings.Split(response, "\n")

	if response == updateStatusErr || len(lines) != okResponseLines {
		return UpdateResult{}, fmt.Errorf("duckdns update failed: %s", response)
	}

	status := lines[0]
	ipv4 := lines[1]
	ipv6 := lines[2]
	updateResult := lines[3]

	if status != updateStatusOk {
		return UpdateResult{}, fmt.Errorf("duckdns unexpected response: %s", response)
	}

	if updateResult == updateResultOk {
		return UpdateResult{UpdaterStatusOk, ipv4, ipv6}, nil
	}

	if updateResult == updateResultUnchanged {
		return UpdateResult{UpdaterStatusUnchanged, ipv4, ipv6}, nil
	}

	return UpdateResult{}, fmt.Errorf("duckdns returned unknown update result: %s", updateResult)
}
