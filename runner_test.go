package action_test

import (
	"context"
	"errors"
	"github.com/neonima/action"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("Should create a new instance of Runner", func(t *testing.T) {
		r := action.New()
		require.NotNil(t, r)
		require.Implements(t, (*action.Runners)(nil), r)
	})
}

func TestRunner_Send(t *testing.T) {
	t.Run("Should send successfully", func(t *testing.T) {
		r := action.New()
		require.NoError(t, r.Start(t.Context()))
		ch := make(chan any)
		r.Send(func() {
			ch <- "hello world"
		})
		res := <-ch
		require.Equal(t, "hello world", res)
	})
}

func TestRunner_Start(t *testing.T) {
	t.Run("Should start successfully", func(t *testing.T) {
		r := action.New()
		require.NoError(t, r.Start(t.Context()))
	})
}

func TestRunner_Done(t *testing.T) {
	t.Run("Should return a channel closed when the runner is stopped", func(t *testing.T) {
		r := action.New()
		ctx, cancel := context.WithCancel(t.Context())
		err := r.Start(ctx)
		require.NoError(t, err)
		cancel()
		<-r.Done()
		require.Error(t, r.Error())
	})
}

func TestRunner_Error(t *testing.T) {
	t.Run("Should return the error of the runner", func(t *testing.T) {
		r := action.New()
		ctx, cancel := context.WithCancel(t.Context())
		err := r.Start(ctx)
		require.NoError(t, err)
		cancel()
	})
}

func TestRunner_WithHook(t *testing.T) {
	t.Run("Should add a hook to the runner", func(t *testing.T) {
		incr := 0
		r := action.New(action.WithHook(func(ctx context.Context) error {
			incr++

			return nil
		}))
		ctx, cancel := context.WithCancel(t.Context())
		require.NoError(t, r.Start(ctx))
		action.Act(r, func() {
			incr--
		})
		cancel()
		<-r.Done()
		require.Error(t, r.Error())
		require.Equal(t, 0, incr)
	})
	t.Run("Should add a hook to the runner and stop the execution if it errored", func(t *testing.T) {
		incr := 0
		r := action.New(action.WithHook(func(ctx context.Context) error {
			return errors.New("error")
		}))
		ctx, cancel := context.WithCancel(t.Context())
		require.NoError(t, r.Start(ctx))
		action.Act(r, func() {
			incr--
		})
		cancel()
		<-r.Done()
		require.Error(t, r.Error())
		require.Equal(t, -1, incr)
	})
}
