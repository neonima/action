package action

import (
	"context"
	"sync/atomic"
	"time"
)

type Runners interface {
	Start(context.Context) error
	Send(Action) (context.Context, context.CancelFunc)
}

type Runner struct {
	stream       chan Action
	isStarted    atomic.Bool
	timeout      time.Duration
	tickDuration time.Duration
	noTick       bool
	ctx          context.Context
}

// New returns a new Runner with default configuration settings.
//
// The default settings are:
//   - tickDuration: 5ms
//   - timeout: 1 minute
//   - stream channel capacity: 1
//   - noTick: true
func New(opts ...func(*Runner)) *Runner {
	r := &Runner{
		stream:       make(chan Action, 1),
		timeout:      time.Minute,
		tickDuration: 5 * time.Millisecond,
		noTick:       true,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithChanSize defines a specific chan size for the actor buffer message queue
// default is 1
func WithChanSize(size int) func(*Runner) {
	return func(r *Runner) {
		r.stream = make(chan Action, size)
	}
}

// WithTickDuration defines the Tick rate duration
func WithTickDuration(duration time.Duration) func(*Runner) {
	return func(r *Runner) {
		r.tickDuration = duration
	}
}

// WithoutTick allows the runner to loop as fast as possible
func WithoutTick() func(*Runner) {
	return func(r *Runner) {
		r.noTick = true
	}
}

func WithTimeout(timeout time.Duration) func(*Runner) {
	return func(r *Runner) {
		r.timeout = timeout
	}
}

// Start startes the runner on a separated goroutine
func (r *Runner) Start(ctx context.Context) error {
	if r.isStarted.Load() {
		return ErrAlreadyStarted
	}
	r.isStarted.Store(true)
	if ctx == nil {
		return ErrNilContext
	}
	r.ctx = ctx
	if r.noTick {
		go r.startWithoutTick(ctx)
		return nil
	}
	go r.start(ctx)
	return nil
}

func (r *Runner) start(ctx context.Context) {
	t := time.NewTicker(10 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-t.C:
		case <-ctx.Done():
			close(r.stream)
			return
		case action, ok := <-r.stream:
			if !ok {
				return
			}
			action()
		}
	}
}

func (r *Runner) startWithoutTick(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(r.stream)
			return
		case action, ok := <-r.stream:
			if !ok {
				return
			}
			action()
		}
	}
}

// Send allow to send the action on the queue
// exported to allow custom implementation, see action.go for examples (which should be enough for most cases)
func (r *Runner) Send(a Action) (context.Context, context.CancelFunc) {
	r.stream <- a
	return context.WithTimeout(r.ctx, r.timeout)
}
