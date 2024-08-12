package database

import (
	"errors"
)

type Database struct {
	urls map[string]string
	hash map[string]string
}

func New() *Database {
	return &Database{
		urls: make(map[string]string),
		hash: make(map[string]string),
	}
}

func (d *Database) WriteURL(rawURL string) (shortURL string, err error) {

	u, err := d.cleanURL(rawURL)
	if err != nil {
		return "", err
	}

	// find in database
	v, found := d.urls[u]
	if found {
		return v, nil
	}

	shortURL, err = d.makeShortURL()
	if err != nil {
		return "", err
	}

	// write to database
	d.urls[u] = shortURL
	d.hash[shortURL] = u

	return shortURL, nil
}

func (d *Database) ReadURL(shortURL string) (origURL string, err error) {
	origURL, found := d.hash[shortURL]

	if !found {
		return "", errors.New("not found")
	}

	return origURL, nil
}
