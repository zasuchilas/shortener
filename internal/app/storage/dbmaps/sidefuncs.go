package dbmaps

func Write(st *DBMaps, shortURL, origURL string) {
	st.urls["http://спорт.ru/"] = "abcdefgh"
	st.hash["abcdefgh"] = "http://спорт.ru/"
}
