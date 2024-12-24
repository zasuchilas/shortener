package envflags

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestTryUseEnvString(t *testing.T) {
	type args struct {
		flagValue string
		envName   string
	}
	tests := []struct {
		name     string
		envValue string
		args     args
		expected string
	}{
		{
			name:     "env is set",
			envValue: "123",
			args: args{
				flagValue: "some_value",
				envName:   "SOME_ENV",
			},
			expected: "123",
		},
		{
			name:     "env isn't set",
			envValue: "",
			args: args{
				flagValue: "some_value",
				envName:   "SOME_ENV",
			},
			expected: "some_value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(tt.args.envName, tt.envValue)
			require.NoError(t, err)

			TryUseEnvString(&tt.args.flagValue, tt.args.envName)
			require.Equal(t, tt.expected, tt.args.flagValue)
		})
	}
}

func TestTryUseEnvBool(t *testing.T) {
	type args struct {
		flagValue bool
		envName   string
	}
	tests := []struct {
		name     string
		envValue string
		args     args
		expected bool
	}{
		{
			name:     "env is set",
			envValue: "true",
			args: args{
				flagValue: false,
				envName:   "SOME_ENV",
			},
			expected: true,
		},
		{
			name:     "env isn't set",
			envValue: "",
			args: args{
				flagValue: true,
				envName:   "SOME_ENV",
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(tt.args.envName, tt.envValue)
			require.NoError(t, err)

			TryUseEnvBool(&tt.args.flagValue, tt.args.envName)
			require.Equal(t, tt.expected, tt.args.flagValue)
		})
	}
}

func TestTryUseEnvBoolPanic(t *testing.T) {
	err := os.Setenv("SOME_ENV", "123true")
	require.NoError(t, err)
	val := false
	require.Panics(t, func() {
		TryUseEnvBool(&val, "SOME_ENV")
	})
}
