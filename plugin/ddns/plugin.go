package ddns

import (
	"context"
	"keeper/pkg/logger"
	"keeper/plugin"
	"keeper/service/ddns"
	"time"
)

const defaultInterval = "10m"

type PluginLoader struct{}

type Config struct {
	Interval       string          `hcl:"interval,optional"`
	DuckDNSConfigs []DuckDNSConfig `hcl:"duckdns,block"`
}

type Plugin struct {
	service *ddns.Service
}

func (*PluginLoader) Name() string {
	return "ddns"
}

func (*PluginLoader) ConfigSchema() any {
	return &Config{Interval: defaultInterval}
}

func (*PluginLoader) Execute(config any) (plugin.Plugin, error) {
	ddnsConfig, err := plugin.CastConfig[Config](config)
	if err != nil {
		return nil, err
	}

	interval, err := time.ParseDuration(ddnsConfig.Interval)
	if err != nil {
		return nil, err
	}

	p := &Plugin{ddns.New(interval)}

	for i := range len(ddnsConfig.DuckDNSConfigs) {
		p.service.AddUpdater(NewDuckDNSUpdater(&ddnsConfig.DuckDNSConfigs[i]))
	}

	return p, nil
}

func (p *Plugin) Run(ctx context.Context) error {
	return p.service.Run(ctx)
}

func (*Plugin) Close() error {
	return nil
}

func init() {
	logger.LogIfErrorTyped[PluginLoader](plugin.Register(&PluginLoader{}))
}
