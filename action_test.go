package action_test

import (
	"errors"
	"github.com/neonima/action"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAct(t *testing.T) {
	type K struct {
		toChange string
	}
	tt := []struct {
		title        string
		data         *K
		expectedData *K
		timeout      time.Duration
	}{
		{
			title:        "Should successfully modify data",
			data:         &K{toChange: "hello"},
			expectedData: &K{toChange: "world"},
			timeout:      time.Minute,
		},
		{
			title:        "Should timeout before sending data",
			data:         &K{toChange: "hello"},
			expectedData: &K{toChange: "hello"},
			timeout:      0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.title, func(t *testing.T) {
			r := action.New(action.WithoutTick(), action.WithTimeout(tc.timeout))
			require.NoError(t, r.Start(t.Context()))
			action.Act(r, func() {
				tc.data.toChange = tc.expectedData.toChange
			})
			require.Equal(t, tc.expectedData.toChange, tc.data.toChange)
		})
	}
}

func TestActGet(t *testing.T) {
	tt := []struct {
		title        string
		data         string
		expectedData string
		timeout      time.Duration
	}{
		{
			title:        "Should successfully modify data",
			data:         "world",
			expectedData: "world",
			timeout:      time.Minute,
		},
		{
			title:        "Should timeout before sending data",
			data:         "world",
			expectedData: "",
			timeout:      0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.title, func(t *testing.T) {
			r := action.New(action.WithoutTick(), action.WithTimeout(tc.timeout))
			require.NoError(t, r.Start(t.Context()))
			res := action.ActGet[string](r, func() string {
				return tc.data
			})
			require.Equal(t, tc.expectedData, res)
		})
	}
}

func TestActErr(t *testing.T) {
	type K struct {
		toChange string
	}
	tt := []struct {
		title        string
		data         *K
		expectedData *K
		err          error
		expectedErr  require.ErrorAssertionFunc
		timeout      time.Duration
	}{
		{
			title:        "Should successfully modify data",
			data:         &K{toChange: "hello"},
			expectedData: &K{toChange: "world"},
			err:          nil,
			expectedErr:  require.NoError,
			timeout:      time.Minute,
		},
		{
			title:        "Should timeout before sending data",
			data:         &K{toChange: "hello"},
			expectedData: &K{toChange: "hello"},
			err:          nil,
			expectedErr:  require.Error,
			timeout:      0,
		},
		{
			title:        "Should return the error if one occurs",
			data:         &K{toChange: "hello"},
			expectedData: &K{toChange: "world"},
			err:          errors.New("error"),
			expectedErr:  require.Error,
			timeout:      time.Minute,
		},
	}

	for _, tc := range tt {
		t.Run(tc.title, func(t *testing.T) {
			r := action.New(action.WithoutTick(), action.WithTimeout(tc.timeout))
			require.NoError(t, r.Start(t.Context()))
			err := action.ActErr(r, func() error {
				tc.data.toChange = tc.expectedData.toChange
				return tc.err
			})
			require.Equal(t, tc.expectedData.toChange, tc.data.toChange)
			tc.expectedErr(t, err)
		})
	}
}

func TestActGet2(t *testing.T) {
	tt := []struct {
		title     string
		a         string
		expectedA string
		b         int
		expectedB int
		timeout   time.Duration
	}{
		{
			title:     "Should successfully modify data",
			a:         "a",
			expectedA: "a",
			b:         100,
			expectedB: 100,
			timeout:   time.Minute,
		},
		{
			title:     "Should timeout before sending data",
			a:         "hello",
			expectedA: "",
			b:         100,
			expectedB: 0,
			timeout:   0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.title, func(t *testing.T) {
			r := action.New(action.WithoutTick(), action.WithTimeout(tc.timeout))
			require.NoError(t, r.Start(t.Context()))
			a, b := action.ActGet2[string, int](r, func() (string, int) {
				return tc.a, tc.b
			})
			require.Equal(t, tc.expectedA, a)
			require.Equal(t, tc.expectedB, b)
		})
	}
}

func TestActGet3(t *testing.T) {
	tt := []struct {
		title     string
		a         string
		expectedA string
		b         int
		expectedB int
		c         float32
		expectedC float32
		timeout   time.Duration
	}{
		{
			title:     "Should successfully modify data",
			a:         "a",
			expectedA: "a",
			b:         100,
			expectedB: 100,
			c:         0.5,
			expectedC: 0.5,
			timeout:   time.Minute,
		},
		{
			title:     "Should timeout before sending data",
			a:         "hello",
			expectedA: "",
			b:         100,
			expectedB: 0,
			c:         0.5,
			expectedC: 0,
			timeout:   0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.title, func(t *testing.T) {
			r := action.New(action.WithoutTick(), action.WithTimeout(tc.timeout))
			require.NoError(t, r.Start(t.Context()))
			a, b, c := action.ActGet3[string, int, float32](r, func() (string, int, float32) {
				return tc.a, tc.b, tc.c
			})
			require.Equal(t, tc.expectedA, a)
			require.Equal(t, tc.expectedB, b)
			require.Equal(t, tc.expectedC, c)
		})
	}
}

func TestActGetErr(t *testing.T) {
	tt := []struct {
		title        string
		data         string
		expectedData string
		err          error
		expectedErr  require.ErrorAssertionFunc
		timeout      time.Duration
	}{
		{
			title:        "Should successfully modify data",
			data:         "world",
			expectedData: "world",
			err:          nil,
			expectedErr:  require.NoError,
			timeout:      time.Minute,
		},
		{
			title:        "Should timeout before sending data",
			data:         "hello",
			expectedData: "",
			err:          nil,
			expectedErr:  require.Error,
			timeout:      0,
		},
		{
			title:        "Should return the error if one occurs",
			data:         "hello",
			expectedData: "",
			err:          errors.New("error"),
			expectedErr:  require.Error,
			timeout:      time.Minute,
		},
	}

	for _, tc := range tt {
		t.Run(tc.title, func(t *testing.T) {
			r := action.New(action.WithoutTick(), action.WithTimeout(tc.timeout))
			require.NoError(t, r.Start(t.Context()))
			res, err := action.ActGetErr[string](r, func() (string, error) {
				if tc.err != nil {
					return "", tc.err
				}
				return tc.data, nil
			})
			require.Equal(t, tc.expectedData, res)
			tc.expectedErr(t, err)
		})
	}
}
