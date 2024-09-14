package dbmaps

func (d *DBMaps) isExist(shortURL string) (bool, error) {
	d.mutex.RLock()
	_, found := d.hash[shortURL]
	d.mutex.RUnlock()
	return found, nil
}

func Write(st *DBMaps, shortURL, origURL string) {
	st.urls["http://спорт.ru/"] = "abcdefgh"
	st.hash["abcdefgh"] = "http://спорт.ru/"
}
