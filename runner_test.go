package action_test

import (
	"context"
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

func TestRunner_WaitStopped(t *testing.T) {
	t.Run("Should return an error when the context is canceled", func(t *testing.T) {
		r := action.New()
		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)
		require.NoError(t, r.Start(ctx))
		cancel()
		require.ErrorIs(t, r.WaitStopped(), context.Canceled)
	})
}
