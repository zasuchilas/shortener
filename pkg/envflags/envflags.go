package envflags

import (
	"log"
	"os"
	"strconv"
)

// TryUseEnvString tries to update the flag values
// from the environment variable (string type).
func TryUseEnvString(flagValue *string, envName string) {
	if env := os.Getenv(envName); env != "" {
		*flagValue = env
	}
}

// TryUseEnvBool tries to update the flag values
// from the environment variable (bool type).
func TryUseEnvBool(flagValue *bool, envName string) {
	if env := os.Getenv(envName); env != "" {
		v, err := strconv.ParseBool(env)
		if err != nil {
			log.Panicf("error trying parse ENV %s (TryUseEnvBool, error: %s)", envName, err.Error())
		}
		*flagValue = v
	}
}
