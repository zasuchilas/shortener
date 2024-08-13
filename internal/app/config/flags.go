package config

import "flag"

var (
	FlagRunAddr string
	FlagOutAddr string
)

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&FlagOutAddr, "b", "localhost:8080", "address and port for include in shortURLs")

	// TODO: implement address validation via the interface flag.Value

	flag.Parse()
}
