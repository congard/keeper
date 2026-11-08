package plugin

import (
	"context"
	"fmt"
)

type Plugin interface {
	Run(ctx context.Context) error
	Close() error
}

type Loader interface {
	Name() string
	ConfigSchema() any
	Execute(config any) (Plugin, error)
}

func CastConfig[C any](config any) (C, error) {
	typedConfig, ok := config.(C)
	if !ok {
		return typedConfig, fmt.Errorf("unexpected config type %T, expected %T", config, typedConfig)
	}
	return typedConfig, nil
}
