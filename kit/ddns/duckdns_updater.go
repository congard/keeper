package ddns

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	url := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&verbose=true",
		url.QueryEscape(updater.domain), url.QueryEscape(updater.token))

	resp, err := updater.client.Get(url)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("duckdns request failed: %w", err)
	}
	defer resp.Body.Close()

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
