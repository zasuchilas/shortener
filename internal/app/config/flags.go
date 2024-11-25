// Package config is used to read the service startup settings.
package config

import (
	"flag"
	"os"
)

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
	flag.StringVar(&SecureFilePath, "s", "./secure.db", "path to the secure data file")
	flag.StringVar(&LogLevel, "l", "info", "logging level")
	flag.Parse()

	// using env (replace)
	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		ServerAddress = envServerAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		BaseURL = envBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		FileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		DatabaseDSN = envDatabaseDSN
	}
	if envSecretKey := os.Getenv("SECRET_KEY"); envSecretKey != "" {
		SecretKey = envSecretKey
	}
	if envSecureFilePath := os.Getenv("SECURE_FILE_PATH"); envSecureFilePath != "" {
		SecureFilePath = envSecureFilePath
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		LogLevel = envLogLevel
	}
}

// TODO: validate flags (net address check with regexp)
