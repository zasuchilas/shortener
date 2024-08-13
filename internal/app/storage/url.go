package storage

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	hostRegex = regexp.MustCompile(`^[a-zA-Z0-9ЁёА-я]+[a-zA-Z0-9ЁёА-я.-]?[a-zA-Z0-9ЁёА-я]+\.[a-zA-ZЁёА-я0-9]{2,}$`)
	// TODO: ru.спорт1abc.рф ru.спорт-1abc.рф ru.спорт.1abc.рф
	// TODO: ru.спорт..1abc.рф ru.спорт.-1abc.рф
)

func (d *Database) cleanURL(raw string) (string, error) {
	// the request body may contain spaces, unlike the query string
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		return "", errors.New("empty URL received")
	}

	// checking the validity of the received url
	// basic conditions
	_, err := url.Parse(raw)
	if err != nil {
		return raw, err
	}

	// checking scheme
	// http2://спорт.рф/ (mailto, ws/wss, tcp, mqtt ...)
	//if u.Scheme != "http" && u.Scheme != "https" {
	//	return raw, errors.New("wrong raw URL (unexpected scheme)")
	//}

	// checking host
	// https://спорт/
	// port
	//if !hostRegex.MatchString(u.Host) {
	//	return raw, errors.New("wrong raw URL (unexpected host)")
	//}

	// checking path
	// https://спорт.рф/ 1/

	// checking query
	// / ? & ;

	// checking fragment
	// #

	// TODO: Chinese URL http://例子.卷筒纸 becomes http://xn--fsqu00a.xn--3lr804guic/.
	//  The xn-- indicates that the character was not originally ASCII

	// TODO: Japanese URL http://example.com/引き割り.html becomes http://example.com/%E5%BC%95%E3%81%8D%E5%89%B2%E3%82%8A.html

	// TODO: omit the scheme or not ? (http/https)

	// TODO: raw or u.String() ?
	return raw, nil
}
