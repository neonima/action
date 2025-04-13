package action

type Action func()
type ActionErr func() error
type ActionReturn[T any] func() T
type ActionReturnWithError[T any] func() (T, error)
type ActionReturn2[A, B any] func() (A, B)
type ActionReturn3[A, B, C any] func() (A, B, C)

type Actioner chan Action

// ActGetErr returns `T` and  an error of the action
func ActGetErr[T any](r Runners, action ActionReturnWithError[T]) (T, error) {
	c := make(chan struct {
		t   T
		err error
	}, 1)

	r.Send(func() {
		t, err := action()
		c <- struct {
			t   T
			err error
		}{t, err}
	})
	ctx := r.Ctx()
	select {
	case <-ctx.Done():
		var t T
		return t, ctx.Err()
	case p := <-c:
		return p.t, p.err
	}
}

// ActErr returns the error of the action
func ActErr(r Runners, action ActionErr) error {
	c := make(chan error, 1)
	r.Send(func() {
		c <- action()
	})
	ctx := r.Ctx()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p := <-c:
		return p
	}
}

// Act only execute the action
func Act(r Runners, action Action) {
	c := make(chan any, 1)
	r.Send(func() {
		action()
		c <- struct{}{}
	})
	ctx := r.Ctx()
	select {
	case <-ctx.Done():
		return
	case <-c:
		return
	}
}

// ActGet returns `T` of the action
func ActGet[T any](r Runners, action ActionReturn[T]) T {
	c := make(chan T)
	r.Send(func() {
		c <- action()
	})
	ctx := r.Ctx()
	select {
	case <-ctx.Done():
		var t T
		return t
	case p := <-c:
		return p
	}
}

// ActGet2 returns `A` and `B` of the action
func ActGet2[A, B any](r Runners, action ActionReturn2[A, B]) (A, B) {
	c := make(chan struct {
		a A
		b B
	}, 1)

	r.Send(func() {
		a, b := action()
		c <- struct {
			a A
			b B
		}{a: a, b: b}
	})
	ctx := r.Ctx()
	select {
	case <-ctx.Done():
		var a A
		var b B
		return a, b
	case p := <-c:
		return p.a, p.b
	}
}

// ActGet3 returns `A`, `B  and `C` of the action
func ActGet3[A, B, C any](r Runners, action ActionReturn3[A, B, C]) (A, B, C) {
	ch := make(chan struct {
		a A
		b B
		c C
	}, 1)
	r.Send(func() {
		a, b, c := action()
		ch <- struct {
			a A
			b B
			c C
		}{a: a, b: b, c: c}
	})
	ctx := r.Ctx()
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
