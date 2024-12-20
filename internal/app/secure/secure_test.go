package secure

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			secure := New(tt.key, "", "")

			encrypted, nonce, err := secure.Encrypt([]byte(tt.data))
			assert.NoError(t, err)

			decrypted, err2 := secure.Decrypt(encrypted, nonce)
			assert.NoError(t, err2)

			assert.Equal(t, tt.data, string(decrypted))
		})
	}
}

func TestPacking(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
	}{
		{
			name:   "positive: test 1",
			userID: 1,
		},
		{
			name:   "positive: test 2",
			userID: 100500,
		},
	}

	key := "supersecretkey"
	secure := New(key, "", "")
	nonce, e := generateRandom(secure.nonceSize)
	assert.NoError(t, e)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hexadecimal := secure.packTokenCookieData(tt.userID, nonce)
			userID, _, err := secure.unpackTokenCookieData(hexadecimal)

			assert.NoError(t, err)
			assert.Equal(t, tt.userID, userID)
		})
	}
}

func BenchmarkSecure_Encrypt(b *testing.B) {
	secure := New("supersecretkey", "", "")
	data := []byte("1234567890")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = secure.Encrypt(data)
	}
}

func BenchmarkSecure_Decrypt(b *testing.B) {
	secure := New("supersecretkey", "", "")
	data := []byte("1234567890")
	encrypted, nonce, _ := secure.Encrypt(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = secure.Decrypt(encrypted, nonce)
	}
}
