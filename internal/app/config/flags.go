// Package config is used to read the service startup settings.
package config

import (
	"flag"
	"github.com/zasuchilas/shortener/pkg/envflags"
)

// Variables
var (
	// ServerAddress is the address and port to run server.
	//   localhost:8080 by default
	ServerAddress string

	// BaseURL is the address and port for include in shortURLs.
	//   localhost:8080 by default
	BaseURL string

	// FileStoragePath is the path to the data storage file.
	//  If you want to use it as a data store, then you need to specify the full path to the storage file,
	//  e.g. ./storage.db (you do not need to create a file, it will be created automatically).
	FileStoragePath string

	// DatabaseDSN is the database connection string.
	//  If you want to use it as a data store, then you need to specify a database connection string,
	//  e.g. host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable
	DatabaseDSN string

	// SecretKey is the secret key for user tokens.
	//  supersecretkey by default
	SecretKey string

	// SecureFilePath is path to the secure data file.
	//  ./secure.db by default
	//
	// This file stores users with their IDs.
	// The latter are used to save the owner of the urls.
	SecureFilePath string

	// LogLevel is logging level in app.
	//  info by default
	LogLevel string

	// EnableHTTPS is enable https flag.
	// false by default
	EnableHTTPS bool
)

// ParseFlags reads the startup flags and environment variables.
//
// If there is an environment variable, then it is used.
// If there is no environment variable, but there is a flag, then a flag is used.
// If there is nothing, the default flag value is used.
func ParseFlags() {
	// using flags (and set default values)
	flag.StringVar(&ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&BaseURL, "b", "localhost:8080", "address and port for include in shortURLs")
	flag.StringVar(&FileStoragePath, "f", "", "path to the data storage file")
	flag.StringVar(&DatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&SecretKey, "k", "supersecretkey", "the secret key for user tokens")
	flag.StringVar(&SecureFilePath, "sec", "./secure.db", "path to the secure data file")
	flag.StringVar(&LogLevel, "l", "info", "logging level")
	flag.BoolVar(&EnableHTTPS, "s", false, "enable https")
	flag.Parse()

	// using env (replace)
	envflags.TryUseEnvString(&ServerAddress, "SERVER_ADDRESS")
	envflags.TryUseEnvString(&BaseURL, "BASE_URL")
	envflags.TryUseEnvString(&FileStoragePath, "FILE_STORAGE_PATH")
	envflags.TryUseEnvString(&DatabaseDSN, "DATABASE_DSN")
	envflags.TryUseEnvString(&SecretKey, "SECRET_KEY")
	envflags.TryUseEnvString(&SecureFilePath, "SECURE_FILE_PATH")
	envflags.TryUseEnvString(&LogLevel, "LOG_LEVEL")
	envflags.TryUseEnvBool(&EnableHTTPS, "ENABLE_HTTPS")
}
