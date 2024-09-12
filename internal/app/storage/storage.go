package storage

type Storage interface {
	WriteURL(rawURL string) (shortURL string, err error)
	ReadURL(shortURL string) (origURL string, err error)
	Self() *Database
}
