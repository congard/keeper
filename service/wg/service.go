package wg

import (
	"context"
	"fmt"
	"keeper/pkg/eventbus"
	"keeper/pkg/svcutil"
	"log/slog"

	"golang.zx2c4.com/wireguard/wgctrl"
)

type ServiceConfig struct {
	device         string
	nameResolver   PeerNameResolver
	watcherOptions []WatcherOption
	context        context.Context
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

	var watcher *Watcher

	logger := slog.With("component", "WgService", "device", config.device)

	if config.context != nil {
		watcher = NewWatcher(config.device, client, append(
			config.watcherOptions, WithPeerNameResolver(config.nameResolver))...)
		svcutil.Go(logger, func() error { return watcher.Run(config.context) })
	}

	return &Service{
		device:       config.device,
		nameResolver: config.nameResolver,
		watcher:      watcher,
		client:       client,
		log:          logger,
	}, nil
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
