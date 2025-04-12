package action

type Action func()
type ActionErr func() error
type ActionReturn[T any] func() T
type ActionReturnWithError[T any] func() (T, error)
type ActionReturn2[A, B any] func() (A, B)
type ActionReturn3[A, B, C any] func() (A, B, C)

type Actioner chan Action

func ActGetErr[T any](r Runners, action ActionReturnWithError[T]) (T, error) {
	c := make(chan struct {
		t   T
		err error
	})

	ctx, cancel := r.Send(func() {
		t, err := action()
		c <- struct {
			t   T
			err error
		}{t, err}
	})
	defer cancel()
	select {
	case <-ctx.Done():
		var t T
		return t, ctx.Err()
	case p := <-c:
		return p.t, p.err
	}
}

func ActErr(r Runners, action ActionErr) error {
	c := make(chan error)
	ctx, cancel := r.Send(func() {
		c <- action()
	})
	defer cancel()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p := <-c:
		return p
	}
}

func Act(r Runners, action Action) {
	c := make(chan any)
	ctx, cancel := r.Send(func() {
		action()
		c <- struct{}{}
	})
	defer cancel()
	select {
	case <-ctx.Done():
		return
	case <-c:
		return
	}
}

func ActGet[T any](r Runners, action ActionReturn[T]) T {
	c := make(chan T)
	ctx, cancel := r.Send(func() {
		c <- action()
	})
	defer cancel()
	select {
	case <-ctx.Done():
		var t T
		return t
	case p := <-c:
		return p
	}
}

func ActGet2[A, B any](r Runners, action ActionReturn2[A, B]) (A, B) {
	c := make(chan struct {
		a A
		b B
	})

	ctx, cancel := r.Send(func() {
		a, b := action()
		c <- struct {
			a A
			b B
		}{a: a, b: b}
	})
	defer cancel()
	select {
	case <-ctx.Done():
		var a A
		var b B
		return a, b
	case p := <-c:
		return p.a, p.b
	}
}

func ActGet3[A, B, C any](r Runners, action ActionReturn3[A, B, C]) (A, B, C) {
	ch := make(chan struct {
		a A
		b B
		c C
	})
	ctx, cancel := r.Send(func() {
		a, b, c := action()
		ch <- struct {
			a A
			b B
			c C
		}{a: a, b: b, c: c}
	})
	defer cancel()
	select {
	case <-ctx.Done():
		var a A
		var b B
		var c C
		return a, b, c
	case p := <-ch:
		return p.a, p.b, p.c
	}
}
