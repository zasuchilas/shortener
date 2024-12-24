package envflags

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTryConfigStringFlag(t *testing.T) {
	type args struct {
		flagValue   string
		configValue string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "flag already set",
			args: args{
				flagValue:   "some_value",
				configValue: "",
			},
			expected: "some_value",
		},
		{
			name: "flag not set",
			args: args{
				flagValue:   "",
				configValue: "123",
			},
			expected: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TryConfigStringFlag(&tt.args.flagValue, tt.args.configValue)
			require.Equal(t, tt.expected, tt.args.flagValue)
		})
	}
}

func TestTryConfigBoolFlag(t *testing.T) {
	type args struct {
		flagValue   bool
		configValue bool
	}
	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "flag already set",
			args: args{
				flagValue:   true,
				configValue: false,
			},
			expected: true,
		},
		{
			name: "flag not set",
			args: args{
				flagValue:   false,
				configValue: false,
			},
			expected: false,
		},
		{
			name: "config is set",
			args: args{
				flagValue:   false,
				configValue: true,
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TryConfigBoolFlag(&tt.args.flagValue, tt.args.configValue)
			require.Equal(t, tt.expected, tt.args.flagValue)
		})
	}
}
