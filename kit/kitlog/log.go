package kitlog

import "log/slog"

type ErrorOption func(*errorOptions)

type errorOptions struct {
	description string
}

func WithDescription(prefix string) ErrorOption {
	return func(lo *errorOptions) {
		lo.description = prefix
	}
}

func Error(err error, opts ...ErrorOption) {
	if err == nil {
		return
	}

	lo := &errorOptions{description: "error"}
	for _, opt := range opts {
		opt(lo)
	}

	slog.Error("an error occurred",
		slog.Any("error", err),
		OptionalString("description", lo.description),
	)
}

func OptionalString(key, val string) slog.Attr {
	if val == "" {
		return slog.Attr{}
	}
	return slog.String(key, val)
}
