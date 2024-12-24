package urlfuncs

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zasuchilas/shortener/internal/app/config"
)

func TestCleanURL(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		expected    string
		expectedErr error
	}{
		{
			name:        "ya.ru",
			raw:         "ya.ru",
			expected:    "ya.ru",
			expectedErr: nil,
		},
		{
			name:        "https://ya.ru",
			raw:         "https://ya.ru",
			expected:    "https://ya.ru",
			expectedErr: nil,
		},
		{
			name:        "httpss://ya.ru",
			raw:         "httpss://ya.ru",
			expected:    "httpss://ya.ru",
			expectedErr: nil,
		},
		{
			name:        "empty",
			raw:         " ",
			expected:    "",
			expectedErr: errors.New("empty URL received"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := CleanURL(tt.raw)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestCleanURLError(t *testing.T) {
	_, err := CleanURL("12http::///http://ya.ru")
	require.ErrorContains(t, err, "first path segment in URL cannot contain colon")
}

func TestEnrichURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		shortURL string
		expected string
	}{
		{
			name:     "localhost:8888 + 19xtf1u2",
			baseURL:  "localhost:8888",
			shortURL: "19xtf1u2",
			expected: "http://localhost:8888/19xtf1u2",
		},
		{
			name:     "http://localhost:8888 + 19xtf1u2",
			baseURL:  "http://localhost:8888",
			shortURL: "19xtf1u2",
			expected: "http://localhost:8888/19xtf1u2",
		},
		{
			name:     "https://localhost:8888 + 19xtf1u2",
			baseURL:  "https://localhost:8888",
			shortURL: "19xtf1u2",
			expected: "https://localhost:8888/19xtf1u2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.BaseURL = tt.baseURL
			actual := EnrichURL(tt.shortURL)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func BenchmarkEnrichURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = EnrichURL("ya.ru")
	}
}

func BenchmarkEnrichURLv2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = EnrichURLv2("ya.ru")
	}
}
