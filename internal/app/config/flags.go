package config

import (
	"flag"
)

var (
	RunAddr string
	OutAddr string

	// TODO: validate flags
	//netAddrRegex = regexp.MustCompile(`^[a-zA-Z0-9ЁёА-я.-]+\.[a-zA-ZЁёА-я0-9]{2,}:[0-9]{3,4}$`)
)

func ParseFlags() {
	flag.StringVar(&RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&OutAddr, "b", "localhost:8080", "address and port for include in shortURLs")
	flag.Parse()
}

//func validateNetAddr(val string) error {
//	if !netAddrRegex.MatchString(val) {
//		return errors.New("")
//	}
//	return nil
//}
