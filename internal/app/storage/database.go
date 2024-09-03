package storage

import (
	"errors"
	"sync"
)

type Database struct {
	Urls  map[string]string
	Hash  map[string]string
	mutex sync.RWMutex
}

func New() Storage {
	return &Database{
		Urls: make(map[string]string),
		Hash: make(map[string]string),
	}
}

func (d *Database) WriteURL(rawURL string) (shortURL string, err error) {

	u, err := d.cleanURL(rawURL)
	if err != nil {
		return "", err
	}

	// find in storage
	v, found := d.Urls[u]
	if found {
		return v, nil
	}

	shortURL, err = d.makeShortURL()
	if err != nil {
		return "", err
	}

	// write to storage
	d.mutex.Lock()
	d.Urls[u] = shortURL
	d.Hash[shortURL] = u
	d.mutex.Unlock()

	return shortURL, nil
}

func (d *Database) ReadURL(shortURL string) (origURL string, err error) {
	d.mutex.RLock()
	origURL, found := d.Hash[shortURL]
	d.mutex.RUnlock()

	if !found {
		return "", errors.New("not found")
	}

	return origURL, nil
}
