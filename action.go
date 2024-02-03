package action

type Action func()
type ActionErr func() error
type ActionReturn[T any] func() T
type ActionReturnWithError[T any] func() (T, error)

type Actioner chan Action

func ActGetErr[T any](r Runners, action ActionReturnWithError[T]) (T, error) {
	c := make(chan struct {
		t   T
		err error
	})

	r.Send(func() {
		t, err := action()
		c <- struct {
			t   T
			err error
		}{t, err}
	})
	p := <-c
	return p.t, p.err

}

func ActGet[T any](r Runners, action ActionReturn[T]) T {
	c := make(chan T)
	r.Send(func() {
		c <- action()
	})
	return <-c

}

func ActErr(r Runners, action ActionErr) error {
	c := make(chan error)
	r.Send(func() {
		c <- action()
	})
	return <-c
}

func Act(r Runners, action Action) {
	r.Send(action)
}
