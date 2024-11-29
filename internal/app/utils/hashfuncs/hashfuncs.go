// Package hashfuncs helps to work with hashes.
package hashfuncs

import (
	"strconv"
)

const (
	// zeroHash contains the first number for calculating url hashes.
	zeroHash int64 = 99999999999

	// heroHash contains the first number for calculating user hashes.
	heroHash int64 = 333333333
)

// EncodeZeroHash calculates the URL hash for the transmitted URL id.
func EncodeZeroHash(id int64) string {
	return encodeHash(id, zeroHash)
}

// DecodeZeroHash calculates the URL id for the transmitted URL hash.
func DecodeZeroHash(hash string) (id int64, err error) {
	return decodeHash(hash, zeroHash)
}

// EncodeHeroHash calculates the user hash for the transmitted user id.
func EncodeHeroHash(id int64) string {
	return encodeHash(id, heroHash)
}

// DecodeHeroHash calculates the user id for the transmitted user hash.
func DecodeHeroHash(hash string) (id int64, err error) {
	return decodeHash(hash, heroHash)
}

func encodeHash(id int64, start int64) string {
	return strconv.FormatInt(start+id, 36)
}

func decodeHash(hash string, start int64) (id int64, err error) {
	i, err := strconv.ParseInt(hash, 36, 64)
	if err != nil {
		return 0, err
	}
	return i - start, nil
}
