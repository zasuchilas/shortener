package hashfuncs

import (
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
	"math/rand"
)

const (
	shortURLLength = 8
	attemptCount   = 10
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
