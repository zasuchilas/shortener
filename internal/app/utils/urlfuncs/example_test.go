package urlfuncs

import (
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/config"
)

func ExampleEnrichURLv2() {
	// The config.BaseURL is installed at the start of the application.
	// Let's install it with our hands for an example.
	config.BaseURL = "localhost:8080"

	out := EnrichURLv2("19xtf1tx")
	fmt.Println(out)

	// Output:
	// http://localhost:8080/19xtf1tx
}
