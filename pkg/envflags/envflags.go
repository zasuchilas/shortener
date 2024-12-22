package envflags

import (
	"log"
	"os"
	"strconv"
	"time"
)

// TryUseEnvString tries to update the flag values
// from the environment variable (string type)
func TryUseEnvString(flagValue *string, envName string) {
	if env := os.Getenv(envName); env != "" {
		*flagValue = env
	}
}

// TryUseEnvDuration tries to update the flag values
// from the environment variable (duration type)
func TryUseEnvDuration(flagValue *time.Duration, envName string) {
	if env := os.Getenv(envName); env != "" {
		v, err := time.ParseDuration(env)
		if err != nil {
			log.Panicf("error trying parse ENV %s (TryUseEnvDuration, error: %s)", envName, err.Error())
		}
		*flagValue = v
	}
}

// TryUseEnvInt tries to update the flag values
// from the environment variable (int type)
func TryUseEnvInt(flagValue *int, envName string) {
	if env := os.Getenv(envName); env != "" {
		v, err := strconv.Atoi(env)
		if err != nil {
			log.Panicf("error trying parse ENV %s (TryUseEnvInt, error: %s)", envName, err.Error())
		}
		*flagValue = v
	}
}

// TryUseEnvBool tries to update the flag values
// from the environment variable (bool type)
func TryUseEnvBool(flagValue *bool, envName string) {
	if env := os.Getenv(envName); env != "" {
		v, err := strconv.ParseBool(env)
		if err != nil {
			log.Panicf("error trying parse ENV %s (TryUseEnvBool, error: %s)", envName, err.Error())
		}
		*flagValue = v
	}
}
