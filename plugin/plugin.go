package plugin

import "context"

type Plugin interface {
	Name() string
	ConfigSchema() any
	Execute(ctx context.Context, config any) error
}
