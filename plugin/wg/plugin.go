package wg

import (
	"keeper/pkg/logger"
	"keeper/plugin"
	"keeper/service/wg"
	"time"
)

type PluginLoader struct{}

type Config struct {
	Device      string            `hcl:"device,label"`
	Interval    string            `hcl:"interval,optional"`
	PeerTimeout string            `hcl:"peer_timeout,optional"`
	PeerNames   map[string]string `hcl:"peer_names,optional"`
}

type Plugin struct {
	*wg.Service
}

func (*PluginLoader) Name() string {
	return "wg"
}

func (*PluginLoader) ConfigSchema() any {
	return &Config{}
}

func (*PluginLoader) Execute(config any) (plugin.Plugin, error) {
	wgConfig, err := plugin.CastConfig[Config](config)
	if err != nil {
		return nil, err
	}

	svcConfig := wg.ServiceConfig{
		Device:         wgConfig.Device,
		WatcherOptions: make([]wg.WatcherOption, 0),
	}

	parseDurationOption := func(value string, factory func(duration time.Duration) wg.WatcherOption) error {
		if len(value) == 0 {
			return nil
		}
		interval, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		svcConfig.WatcherOptions = append(svcConfig.WatcherOptions, factory(interval))
		return nil
	}

	if err := parseDurationOption(wgConfig.Interval, wg.WithInterval); err != nil {
		return nil, err
	}

	if err := parseDurationOption(wgConfig.PeerTimeout, wg.WithPeerTimeout); err != nil {
		return nil, err
	}

	if wgConfig.PeerNames != nil {
		svcConfig.NameResolver = func(pubkey string) (name string) {
			if name, exists := wgConfig.PeerNames[pubkey]; exists {
				return name
			}
			return wg.DefaultPeerNameResolver(pubkey)
		}
	}

	service, err := wg.NewService(svcConfig)
	if err != nil {
		return nil, err
	}

	return &Plugin{service}, nil
}

func init() {
	logger.LogIfErrorTyped[PluginLoader](plugin.Register(&PluginLoader{}))
}
