package action

import (
	"context"
	"sync/atomic"
	"time"
)

type Runners interface {
	Start(context.Context) *Runner
	Send(Action)
}

type Runner struct {
	stream    chan Action
	isStarted atomic.Bool
	timeout   time.Duration
}

func NewRunner(opts ...func(*Runner)) *Runner {
	r := &Runner{
		stream:  make(chan Action, 20),
		timeout: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithChanSize(size int) func(*Runner) {
	return func(r *Runner) {
		r.stream = make(chan Action, size)
	}
}

func (r *Runner) Start(ctx context.Context) *Runner {
	if r.isStarted.Load() {
		return r
	}
	r.isStarted.Store(true)
	if ctx == nil {
		ctx = context.Background()
	}
	go func() {
		t := time.NewTicker(10 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				close(r.stream)
				return
			case <-t.C:
			case action, ok := <-r.stream:
				if !ok {
					return
				}
				action()
			}
		}
	}()
	return r
}

func (r *Runner) Send(a Action) {
	r.stream <- a
}
