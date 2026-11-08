package ddns

import "keeper/service/ddns"

type DuckDNSConfig struct {
	Domain string `hcl:"domain,attr"`
	Token  string `hcl:"token,attr"`
	BindIP string `hcl:"bind_ip,attr"`
}

func NewDuckDNSUpdater(config *DuckDNSConfig) *ddns.DuckDNSUpdater {
	return ddns.NewDuckDNSUpdater(config.Domain, config.Token, config.BindIP)
}
