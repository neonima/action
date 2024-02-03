package action

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Actable[T any] struct {
	Value  T
	runner Runners
}

func NewActable[T any](value T, opts ...func(asyncable Actable[T])) Actable[T] {
	r := NewRunner()
	r.Start(context.Background())
	a := Actable[T]{
		Value:  value,
		runner: r,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(a)
	}
	return a
}

func WithRunner[T any](ctx context.Context, runner Runners) func(*Actable[T]) {
	if runner == nil {
		return nil
	}

	runner.Start(ctx)
	return func(a *Actable[T]) {
		a.runner = runner
	}
}

func WithDecorator[T any]() func(*Actable[T]) {
	return func(a *Actable[T]) {}
}

func (a *Actable[T]) Set(value T) {
	a.usePackageRunnerIfNoRunner()
	Act(a.runner, func() {
		a.Value = value
	})
}

func (a *Actable[T]) Get() T {
	a.usePackageRunnerIfNoRunner()
	return ActGet(a.runner, func() T { return a.Value })
}

func (a *Actable[T]) UnmarshalJSON(b []byte) error {
	var obj any
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	if reflect.TypeOf(a.Value) != reflect.TypeOf(obj) {
		return fmt.Errorf("trying to unmarshal type: %T to type %T", a.Value, obj)
	}

	a.Value = obj.(T)
	return nil
}

func (a *Actable[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(&a.Value)
}

func (a *Actable[T]) UnmarshalYAML(n *yaml.Node) error {
	var obj any
	if err := n.Decode(&obj); err != nil {
		return err
	}

	if reflect.TypeOf(a.Value) != reflect.TypeOf(obj) {
		return fmt.Errorf("trying to unmarshal type: %T to type %T", a.Value, obj)
	}

	a.Value = obj.(T)
	return nil
}

func (a *Actable[T]) MarshalYAML() (any, error) {
	return yaml.Marshal(&a.Value)
}

var unspecifiedRunner Runners

func (a *Actable[T]) usePackageRunnerIfNoRunner() {
	if a.runner != nil {
		return
	}
	if unspecifiedRunner == nil {
		unspecifiedRunner = NewRunner(WithChanSize(1))
	}
	a.runner = unspecifiedRunner
	a.runner.Start(context.Background())
}
