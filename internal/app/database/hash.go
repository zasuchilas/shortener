package database

import (
	"fmt"
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

func (d *Database) makeShortURL() (shortURL string, err error) {

	for i := 0; i < attemptCount; i++ {
		shortURL = randStringRunes(shortURLLength)

		// check is already used
		_, found := d.hash[shortURL]
		if !found {
			break
		}
		shortURL = ""
	}

	if shortURL == "" {
		err = fmt.Errorf("failed to generate a short URL, used %d attempts", attemptCount)
	}

	return shortURL, err
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
