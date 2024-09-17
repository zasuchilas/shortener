package hashfuncs

import (
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
)

const (
	shortURLLength       = 8
	attemptCount         = 10
	zeroHash       int64 = 99999999999
)

func init() {
	// TODO: deprecated
	//rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func MakeRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func MakeShortURLCandidate() string {
	return MakeRandomString(shortURLLength)
}

func MakeShortURL(isExist func(string) (bool, error)) (shortURL string, err error) {

	for i := 0; i < attemptCount; i++ {
		shortURL = MakeRandomString(shortURLLength)

		// check is already used
		if found, e := isExist(shortURL); !found {
			logger.Log.Info("error in isExist", zap.Error(e))
			break
		}
		shortURL = ""
	}

	if shortURL == "" {
		err = fmt.Errorf("failed to generate a short URL, used %d attempts", attemptCount)
	}

	return shortURL, err
}

func EncodeZeroHash(id int64) string {
	return strconv.FormatInt(zeroHash+id, 36)
}

func DecodeZeroHash(hash string) (id int64, err error) {
	i, err := strconv.ParseInt(hash, 36, 64)
	if err != nil {
		return 0, err
	}
	return i - zeroHash, nil
}
