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
		//{
		//	name: "-1000",
		//	value: -1000,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := EncodeZeroHash(tt.value)
			id, err := DecodeZeroHash(hash)
			t.Logf("value: %d hash: %s id: %d value==id: %v len(hash): %d",
				tt.value, hash, id, tt.value == id, len(hash))

			assert.Equal(t, tt.value, id)

			assert.NoError(t, err)

			assert.Equal(t, 8, len(hash))
		})
	}

}
