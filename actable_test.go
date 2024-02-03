package action_test

import (
	"encoding/json"
	"testing"

	"github.com/neonima/action"
	"github.com/stretchr/testify/require"
)

func TestAsyncable_UnmarshalJSON(t *testing.T) {
	type testCase[T any] struct {
		name      string
		inputStr  string
		expected  string
		Input     action.Actable[string] `yaml:"input"`
		assertErr require.ErrorAssertionFunc
	}
	tests := []testCase[string]{
		{
			name:      "pouette",
			inputStr:  `{"input":"bouya", "Async":"yep yep"}`,
			assertErr: require.NoError,
			expected:  "bouya",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.inputStr), &tt)
			tt.assertErr(t, err)
			require.Equal(t, tt.expected, tt.Input.Get())
		})
	}
}
