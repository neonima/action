package action

import (
	"context"
	"sync"
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
	hooks     []func(context.Context) error
	done      chan struct{}
	err       atomic.Pointer[error]
	sync.Once
}

// New returns a new Runner with default configuration settings.
//
// The default settings are:
//   - stream channel capacity: 1
func New(opts ...func(*Runner)) *Runner {
	r := &Runner{
		stream: make(chan Action, 1),
		done:   make(chan struct{}, 1),
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

// WithHook adds a hook to the runner.
// The hook will be called after each action is executed.
// The hook can be used to perform some cleanup or logging.
func WithHook(h func(ctx context.Context) error) func(*Runner) {
	return func(r *Runner) {
		if h == nil {
			return
		}

		r.hooks = append(r.hooks, h)
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
	defer func() {
		r.Once.Do(func() {
			close(r.stream)
			close(r.done)
		})
	}()
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			r.err.Store(&err)
			return
		case action, ok := <-r.stream:
			if !ok {
				return
			}
			action()
			for _, h := range r.hooks {
				if err := h(ctx); err != nil {
					r.err.Store(&err)
					return
				}
			}
		}
	}
}

// Done returns a channel closed when the runner is stopped.
func (r *Runner) Done() <-chan struct{} {
	return r.done
}

// Err returns the error of the runner. To be used with Done()
func (r *Runner) Error() error {
	errPtr := r.err.Load()
	if errPtr == nil {
		return nil
	}
	return *errPtr
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
