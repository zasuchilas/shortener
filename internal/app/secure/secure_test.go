package secure

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSecure(t *testing.T) {
	tests := []struct {
		name string
		key  string
		data string
	}{
		{
			name: "positive: test 1",
			key:  "supersecretkey",
			data: "1234567890",
		},
		{
			name: "positive: test 2",
			key:  "supersecretkey",
			data: "ABCabc123_#$",
		},
		{
			name: "positive: small key < 32 byte",
			key:  "key",
			data: "1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secure := New(tt.key)

			encrypted, nonce, err := secure.Encrypt([]byte(tt.data))
			assert.NoError(t, err)

			decrypted, err2 := secure.Decrypt(encrypted, nonce)
			assert.NoError(t, err2)

			assert.Equal(t, tt.data, string(decrypted))
		})
	}
}
