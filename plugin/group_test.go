package plugin

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"
)

// Success returns at once, no retry and no delay
func TestRetryRunner_Success(t *testing.T) {
	calls := 0
	runner := &RetryRunner{Delay: time.Hour}

	plugin := &fakePlugin{runFn: func(context.Context) error {
		calls++
		return nil
	}}

	if err := runner.Run(context.Background(), plugin); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

// A zero MaxAttempts means unlimited retries
func TestRetryRunner_RetriesUntilSuccess(t *testing.T) {
	calls := 0
	runner := &RetryRunner{MaxAttempts: 0}

	plugin := &fakePlugin{runFn: func(context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	}}

	if err := runner.Run(context.Background(), plugin); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

// Stops at MaxAttempts and returns the wrapped error
func TestRetryRunner_ExhaustsAttempts(t *testing.T) {
	errBoom := errors.New("boom")
	calls := 0
	runner := &RetryRunner{MaxAttempts: 3}

	plugin := &fakePlugin{runFn: func(context.Context) error {
		calls++
		return errBoom
	}}

	err := runner.Run(context.Background(), plugin)
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected error to wrap %v, got %v", errBoom, err)
	}
}

// Graceful errors stop the loop and are returned as nil
func TestRetryRunner_GracefulStop(t *testing.T) {
	calls := 0
	runner := &RetryRunner{MaxAttempts: 3}

	plugin := &fakePlugin{runFn: func(context.Context) error {
		calls++
		return context.Canceled
	}}

	if err := runner.Run(context.Background(), plugin); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

// A panic becomes an error and is retried
func TestRetryRunner_Panic(t *testing.T) {
	calls := 0
	runner := &RetryRunner{MaxAttempts: 2}

	plugin := &fakePlugin{runFn: func(context.Context) error {
		calls++
		panic("kaboom")
	}}

	err := runner.Run(context.Background(), plugin)
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	if err == nil {
		t.Fatal("expected the panic to become an error")
	}
}

type fakePlugin struct {
	runFn   func(ctx context.Context) error
	closeFn func() error
}

func (p *fakePlugin) Run(ctx context.Context) error {
	if p.runFn == nil {
		return nil
	}
	return p.runFn(ctx)
}

func (p *fakePlugin) Close() error {
	if p.closeFn == nil {
		return nil
	}
	return p.closeFn()
}

// Both plugins must run at the same time
func TestGroupRun_Concurrency(t *testing.T) {
	group := NewGroup(GroupConfig{})

	var started sync.WaitGroup
	started.Add(2)
	allStarted := make(chan struct{})

	go func() {
		started.Wait()
		close(allStarted)
	}()

	runFn := func(context.Context) error {
		started.Done()
		select {
		case <-allStarted:
			return nil
		case <-time.After(2 * time.Second):
			return errors.New("other plugin did not start")
		}
	}

	group.Add(&fakePlugin{runFn: runFn}, &OnceRunner{})
	group.Add(&fakePlugin{runFn: runFn}, &OnceRunner{})

	if err := group.Run(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// Each failure is reported and returned by Run
func TestGroupRun_Failures(t *testing.T) {
	errA := errors.New("a failed")
	errB := errors.New("b failed")

	var mu sync.Mutex
	var reported []error

	group := NewGroup(GroupConfig{
		FailHandler: func(_ Plugin, err error) {
			mu.Lock()
			defer mu.Unlock()
			reported = append(reported, err)
		},
	})

	group.Add(&fakePlugin{runFn: func(context.Context) error { return errA }}, &OnceRunner{})
	group.Add(&fakePlugin{runFn: func(context.Context) error { return errB }}, &OnceRunner{})

	err := group.Run(context.Background())

	mu.Lock()
	count := len(reported)
	mu.Unlock()

	if count != 2 {
		t.Fatalf("expected fail handler to be called twice, got %d", count)
	}
	if !errors.Is(err, errA) {
		t.Errorf("expected returned error to wrap %v, got %v", errA, err)
	}
	if !errors.Is(err, errB) {
		t.Errorf("expected returned error to wrap %v, got %v", errB, err)
	}
}

// Graceful stops are not reported and give no error
func TestGroupRun_GracefulStops(t *testing.T) {
	var mu sync.Mutex
	called := 0

	group := NewGroup(GroupConfig{
		FailHandler: func(Plugin, error) {
			mu.Lock()
			defer mu.Unlock()
			called++
		},
	})

	group.Add(&fakePlugin{runFn: func(context.Context) error { return nil }}, &OnceRunner{})
	group.Add(&fakePlugin{runFn: func(context.Context) error { return context.Canceled }}, &OnceRunner{})
	group.Add(&fakePlugin{runFn: func(context.Context) error { return context.DeadlineExceeded }}, &OnceRunner{})

	if err := group.Run(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if called != 0 {
		t.Fatalf("expected fail handler not to be called, got %d call(s)", called)
	}
}

// Plugins are closed in the reverse order they were added
func TestGroupClose_ReverseOrder(t *testing.T) {
	group := NewGroup(GroupConfig{})
	tags := []string{"a", "b", "c"}

	var closed []string
	for _, name := range tags {
		group.Add(&fakePlugin{closeFn: func() error {
			closed = append(closed, name)
			return nil
		}}, &OnceRunner{})
	}

	if err := group.Close(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := slices.Clone(tags)
	slices.Reverse(want)
	if !slices.Equal(closed, want) {
		t.Fatalf("expected close order %v, got %v", want, closed)
	}
}
