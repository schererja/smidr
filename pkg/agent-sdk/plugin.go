package sdk

import "context"

type Plugin interface {
	// Name returns the name of the plugin.
	Name() string
	Version() string
	Description() string
	Initialize(
		ctx context.Context,
		config map[string]any,
	) error
	Shutdown(ctx context.Context) error
}
