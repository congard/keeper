package svcutil

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"runtime"
	"runtime/debug"
)

// Go runs f in a new goroutine and logs how it ended:
//   - clean exit: not logged;
//   - context.Canceled / context.DeadlineExceeded: logged at Info level;
//   - any other error: logged at Error level;
//   - panic: recovered and logged at Error level instead of crashing the process.
func Go(logger *slog.Logger, f func() error) {
	if f == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panicked", "function", name, "panic", r, "stack", string(debug.Stack()))
			}
		}()

		err := f()
		switch {
		case err == nil:
			// exited cleanly, nothing to log
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			logger.Info("shutdown", "function", name)
		default:
			logger.Error("stopped unexpectedly", "function", name, "error", err)
		}
	}()
}
