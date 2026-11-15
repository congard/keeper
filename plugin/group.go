package plugin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Runner interface {
	Run(ctx context.Context, plugin Plugin) error
}

type OnceRunner struct{}

type RetryRunner struct {
	MaxAttempts int
	Delay       time.Duration
}

func recoveryFunc(p Plugin, err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("%T panicked: %v", p, r)
	}
}

func (*OnceRunner) Run(ctx context.Context, plugin Plugin) (err error) {
	defer recoveryFunc(plugin, &err)
	return plugin.Run(ctx)
}

func (r *RetryRunner) Run(ctx context.Context, plugin Plugin) error {
	attemptCount := 0

	tryRun := func() (tryAgain bool, err error) {
		attemptCount++

		defer func() {
			if isGracefulStop(err) {
				tryAgain = false
				err = nil
				return
			}

			if attemptCount == r.MaxAttempts {
				tryAgain = false
				err = fmt.Errorf("%T failed after %d attempt(s): %w", plugin, attemptCount, err)
				return
			}

			tryAgain = true
		}()

		defer recoveryFunc(plugin, &err)

		err = plugin.Run(ctx)
		return
	}

	for {
		tryAgain, err := tryRun()
		if !tryAgain {
			return err
		}

		if r.Delay > 0 {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(r.Delay):
			}
		}
	}
}

func DefaultRunner() Runner {
	return &RetryRunner{MaxAttempts: 0, Delay: 10 * time.Second}
}

type FailHandler func(plugin Plugin, err error)

func DefaultFailHandler(plugin Plugin, err error) {
	attributes := []any{slog.String("plugin", fmt.Sprintf("%T", plugin))}

	switch {
	case isGracefulStop(err):
		slog.Info("plugin stopped", attributes...)
	default:
		slog.Error("plugin failed", append(attributes, slog.Any("error", err))...)
	}
}

func isGracefulStop(err error) bool {
	return err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

type entry struct {
	plugin Plugin
	runner Runner
}

type Group struct {
	entries     []entry
	failHandler FailHandler
}

type GroupConfig struct {
	FailHandler FailHandler
}

func NewGroup(config GroupConfig) *Group {
	if config.FailHandler == nil {
		config.FailHandler = DefaultFailHandler
	}
	return &Group{failHandler: config.FailHandler}
}

func (g *Group) Add(plugin Plugin, runner Runner) {
	g.entries = append(g.entries, entry{plugin, runner})
}

func (g *Group) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	errs := make([]error, len(g.entries))

	for i, e := range g.entries {
		wg.Go(func() {
			err := e.runner.Run(ctx, e.plugin)
			if !isGracefulStop(err) {
				g.failHandler(e.plugin, err)
				errs[i] = fmt.Errorf("%T stopped: %w", e.plugin, err)
			}
		})
	}

	wg.Wait()
	return errors.Join(errs...)
}

func (g *Group) Close() error {
	var errs []error
	for i := range len(g.entries) {
		plugin := g.entries[len(g.entries)-i-1].plugin
		errs = append(errs, plugin.Close())
	}
	return errors.Join(errs...)
}
