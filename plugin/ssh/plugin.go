package ssh

import (
	"context"
	"keeper/pkg/logger"
	"keeper/plugin"
	"keeper/service/ssh"
)

type WatcherPluginLoader struct{}
type WatcherConfig struct{}

type WatcherPlugin struct {
	watcher *ssh.Watcher
}

func (*WatcherPluginLoader) Name() string {
	return "ssh_watcher"
}

func (*WatcherPluginLoader) ConfigSchema() any {
	return &WatcherConfig{}
}

func (*WatcherPluginLoader) Execute(config any) (plugin.Plugin, error) {
	if _, err := plugin.CastConfig[WatcherConfig](config); err != nil {
		return nil, err
	}

	return &WatcherPlugin{ssh.NewWatcher()}, nil
}

func (p *WatcherPlugin) Run(ctx context.Context) error {
	return p.watcher.Run(ctx)
}

func (*WatcherPlugin) Close() error {
	return nil
}

func init() {
	logger.LogIfErrorTyped[WatcherPluginLoader](plugin.Register(&WatcherPluginLoader{}))
}
