package action_test

import (
	"github.com/neonima/action"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	r := action.New()
	require.NotNil(t, r)
	require.Implements(t, (*action.Runners)(nil), r)
}

func TestRunner_Send(t *testing.T) {
	r := action.New()
	require.NoError(t, r.Start(t.Context()))
	ch := make(chan any)
	r.Send(func() {
		ch <- "hello world"
	})
	res := <-ch
	require.Equal(t, "hello world", res)
}

func TestRunner_Start(t *testing.T) {
	r := action.New()
	require.NoError(t, r.Start(t.Context()))
}
