package stats

import (
	"context"
	"keeper/pkg/logger"
	"keeper/plugin"
	"keeper/service/stats"
	"time"
)

const defaultInterval = "5m"

type PluginLoader struct{}

type Config struct {
	Interval string `hcl:"interval,optional"`
}

type Plugin struct {
	service *stats.Service
}

func (*PluginLoader) Name() string {
	return "stats"
}

func (*PluginLoader) ConfigSchema() any {
	return &Config{Interval: defaultInterval}
}

func (*PluginLoader) Execute(config any) (plugin.Plugin, error) {
	statsConfig, err := plugin.CastConfig[Config](config)
	if err != nil {
		return nil, err
	}

	interval, err := time.ParseDuration(statsConfig.Interval)
	if err != nil {
		return nil, err
	}

	return &Plugin{stats.NewService(interval)}, nil
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
