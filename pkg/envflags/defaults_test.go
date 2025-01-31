package envflags

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTryDefaultStringFlag(t *testing.T) {
	type args struct {
		flagValue    string
		defaultValue string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "flag already set",
			args: args{
				flagValue:    "some_value",
				defaultValue: "",
			},
			expected: "some_value",
		},
		{
			name: "flag not set",
			args: args{
				flagValue:    "",
				defaultValue: "123",
			},
			expected: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TryDefaultStringFlag(&tt.args.flagValue, tt.args.defaultValue)
			require.Equal(t, tt.expected, tt.args.flagValue)
		})
	}
}

func TestTryDefaultBoolFlag(t *testing.T) {
	type args struct {
		flagValue    bool
		defaultValue bool
	}
	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "flag already set",
			args: args{
				flagValue:    true,
				defaultValue: false,
			},
			expected: true,
		},
		{
			name: "flag not set",
			args: args{
				flagValue:    false,
				defaultValue: false,
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TryDefaultBoolFlag(&tt.args.flagValue, tt.args.defaultValue)
			require.Equal(t, tt.expected, tt.args.flagValue)
		})
	}
}
