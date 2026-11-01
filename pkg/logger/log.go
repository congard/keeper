package logger

import (
	"log/slog"
	"reflect"
	"slices"
)

type ErrorOption func(*errorOptions)

type errorOptions struct {
	description string
}

func WithDescription(prefix string) slog.Attr {
	return OptionalString("description", prefix)
}

func logError(err error, args []any) {
	slog.Error("an error occurred", append(args, slog.Any("error", err))...)
}

func LogIfError(err error, args ...any) {
	if err == nil {
		return
	}
	logError(err, slices.Clone(args))
}

func LogIfErrorTyped[T any](err error, args ...any) {
	if err == nil {
		return
	}
	logError(err, append(slices.Clone(args), slog.String("type", reflect.TypeFor[T]().String())))
}

func OptionalString(key, val string) slog.Attr {
	if val == "" {
		return slog.Attr{}
	}
	return slog.String(key, val)
}

func ForComponent[T any](args ...any) *slog.Logger {
	return slog.With(append(slices.Clone(args), slog.String("component", reflect.TypeFor[T]().String()))...)
}
