package hashfuncs

import (
	"strconv"
)

const (
	zeroHash int64 = 99999999999
	heroHash int64 = 333333333
)

func EncodeZeroHash(id int64) string {
	return encodeHash(id, zeroHash)
}

func DecodeZeroHash(hash string) (id int64, err error) {
	return decodeHash(hash, zeroHash)
}

func EncodeHeroHash(id int64) string {
	return encodeHash(id, heroHash)
}

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
