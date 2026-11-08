package wg

import (
	"context"
	"fmt"
	"keeper/pkg/eventbus"
	"keeper/pkg/logger"
	"log/slog"

	"golang.zx2c4.com/wireguard/wgctrl"
)

type ServiceConfig struct {
	Device         string
	NameResolver   PeerNameResolver
	WatcherOptions []WatcherOption
}

type Service struct {
	device       string
	nameResolver PeerNameResolver
	watcher      *Watcher
	client       *wgctrl.Client
	log          *slog.Logger
}

func NewService(config ServiceConfig) (*Service, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to open wgctrl client: %w", err)
	}

	if config.NameResolver == nil {
		config.NameResolver = DefaultPeerNameResolver
	}

	watcher := NewWatcher(config.Device, client, append(
		config.WatcherOptions, WithPeerNameResolver(config.NameResolver))...)

	return &Service{
		device:       config.Device,
		nameResolver: config.NameResolver,
		watcher:      watcher,
		client:       client,
		log:          logger.ForComponent[Service]("device", config.Device),
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	return s.watcher.Run(ctx)
}

func (s *Service) Close() error {
	return s.client.Close()
}

func (s *Service) EventBroadcaster() eventbus.Broadcaster[Event] {
	return s.watcher.EventBroadcaster()
}

func (s *Service) Peers() ([]Peer, error) {
	dev, err := s.client.Device(s.device)
	if err != nil {
		s.log.Error("fetching error", "error", err)
		return nil, err
	}

	peers := make([]Peer, 0, len(dev.Peers))
	for _, rawPeer := range dev.Peers {
		peers = append(peers, newPeer(rawPeer, s.nameResolver))
	}

	return peers, nil
}
