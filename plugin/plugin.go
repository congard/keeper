package plugin

import "context"

type Plugin interface {
	Name() string
	NewConfigStruct() any
	Execute(config any, ctx context.Context)
}
