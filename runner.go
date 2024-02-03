package action

import (
	"context"
	"sync/atomic"
)

type Runners interface {
	Start(context.Context)
	Send(Action)
	Stop()
}

type Runner struct {
	stream    chan Action
	isStarted atomic.Bool
}

func NewRunner(opts ...func(*Runner)) Runners {
	r := &Runner{
		stream: make(chan Action, 20),
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

func (r *Runner) Start(ctx context.Context) {
	if r.isStarted.Load() {
		return
	}
	r.isStarted.Store(true)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case action, ok := <-r.stream:
				if !ok {
					return
				}
				action()
			}
		}
	}()
}

func (r *Runner) Stop() {
	close(r.stream)
}

func (r *Runner) Send(a Action) {
	r.stream <- a
}
