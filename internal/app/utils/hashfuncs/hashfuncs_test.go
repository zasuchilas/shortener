package hashfuncs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZeroHash(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{
			name:  "zero",
			value: zeroHash,
		},
		{
			name:  "1",
			value: 1,
		},
		{
			name:  "10000000",
			value: 10000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := EncodeZeroHash(tt.value)
			id, err := DecodeZeroHash(hash)

			//t.Logf("value: %d hash: %s id: %d value==id: %v len(hash): %d",
			//	tt.value, hash, id, tt.value == id, len(hash))

			assert.Equal(t, tt.value, id)

			assert.NoError(t, err)

			assert.Equal(t, 8, len(hash))
		})
	}

}

func TestHeroHash(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{
			name:  "hero",
			value: heroHash,
		},
		{
			name:  "1",
			value: 1,
		},
		{
			name:  "10000000",
			value: 10000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := EncodeHeroHash(tt.value)
			id, err := DecodeHeroHash(hash)

			assert.Equal(t, tt.value, id)

			assert.NoError(t, err)

			assert.Equal(t, 6, len(hash))
		})
	}
}

func Test_encodeHash(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		result string
	}{
		{
			name:   "1",
			value:  1,
			result: "5ighna",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := encodeHash(tt.value, heroHash)

			assert.Equal(t, tt.result, hash)

		})
	}

}

func Test_decodeHash(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		result  int64
		withErr bool
	}{
		{
			name:    "1",
			value:   "5ighna",
			result:  1,
			withErr: false,
		},
		{
			name:    "err",
			value:   "5ighnawwwwwwwwwwwwwwwwww",
			result:  0,
			withErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := decodeHash(tt.value, heroHash)

			if !tt.withErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.result, hash)
		})
	}
}
