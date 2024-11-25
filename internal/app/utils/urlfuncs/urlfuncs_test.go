package urlfuncs

import (
	"errors"
	"fmt"
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

func ExampleEnrichURLv2() {
	// The config.BaseURL is installed at the start of the application.
	// Let's install it with our hands for an example.
	config.BaseURL = "localhost:8080"

	out := EnrichURLv2("19xtf1tx")
	fmt.Println(out)

	// Output:
	// http://localhost:8080/19xtf1tx
}
