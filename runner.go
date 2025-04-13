package action

import (
	"context"
	"sync/atomic"
)

type Runners interface {
	Start(context.Context) error
	Send(Action)
	Ctx() context.Context
}

type Runner struct {
	stream    chan Action
	isStarted atomic.Bool
	ctx       context.Context
	errCh     chan error
}

// New returns a new Runner with default configuration settings.
//
// The default settings are:
//   - stream channel capacity: 1
func New(opts ...func(*Runner)) *Runner {
	r := &Runner{
		stream: make(chan Action, 1),
		errCh:  make(chan error, 1),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithChanSize defines a specific chan size for the actor buffer message queue
// default is 1. If 0, will be ignored.
func WithChanSize(size int) func(*Runner) {
	return func(r *Runner) {
		if size <= 0 {
			return
		}
		r.stream = make(chan Action, size)
	}
}

// Start starts the runner on a separated goroutine
func (r *Runner) Start(ctx context.Context) error {
	if r.isStarted.Load() {
		return ErrAlreadyStarted
	}
	r.isStarted.Store(true)
	if ctx == nil {
		return ErrNilContext
	}
	r.ctx = ctx
	go r.start(ctx)
	return nil
}

func (r *Runner) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(r.stream)
			r.errCh <- ctx.Err()
			return
		case action, ok := <-r.stream:
			if !ok {
				return
			}
			action()
		}
	}
}

// WaitStopped blocks until the runner shuts down
// and returns the shutdown reason, typically context.Canceled.
func (r *Runner) WaitStopped() error {
	return <-r.errCh
}

// Send enqueues an action onto the actor's queue.
// It is exported to support custom implementations, but direct use is discouraged. See action.go for examples, which should suffice in most cases.
func (r *Runner) Send(a Action) {
	r.stream <- a
}

// Ctx returns the context
func (r *Runner) Ctx() context.Context {
	return r.ctx
}
