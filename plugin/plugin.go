package plugin

import (
	"context"
	"fmt"
)

type Plugin interface {
	Name() string
	ConfigSchema() any
	Execute(ctx context.Context, config any) error
}

func CastConfig[C any](config any) (C, error) {
	typedConfig, ok := config.(C)
	if !ok {
		return typedConfig, fmt.Errorf("unexpected config type %T, expected %T", config, typedConfig)
	}
	return typedConfig, nil
}
