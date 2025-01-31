// Package config is used to read the service startup settings.
package config

import (
	"flag"
	"log"

	"github.com/zasuchilas/shortener/pkg/envflags"
)

// Variables
var (
	// ServerAddress is the address and port to run server.
	ServerAddress        string
	defaultServerAddress = "localhost:8080"

	// GRPCServerAddress is the address and port to run gRPC server.
	GRPCServerAddress        string
	defaultGRPCServerAddress = "localhost:9080"

	// BaseURL is the address and port for include in shortURLs.
	BaseURL        string
	defaultBaseURL = "localhost:8080"

	// FileStoragePath is the path to the data storage file.
	//  If you want to use it as a data store, then you need to specify the full path to the storage file,
	//  e.g. ./storage.db (you do not need to create a file, it will be created automatically).
	FileStoragePath        string
	defaultFileStoragePath = ""

	// DatabaseDSN is the database connection string.
	//  If you want to use it as a data store, then you need to specify a database connection string,
	//  e.g. host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable
	DatabaseDSN        string
	defaultDatabaseDSN = ""

	// SecretKey is the secret key for user tokens.
	//  supersecretkey by default
	SecretKey        string
	defaultSecretKey = "supersecretkey"

	// SecureFilePath is path to the secure data file.
	//
	// This file stores users with their IDs.
	// The latter are used to save the owner of the urls.
	SecureFilePath        string
	defaultSecureFilePath = "./secure.db"

	// LogLevel is logging level in app.
	LogLevel        string
	defaultLogLevel = "info"

	// EnableHTTPS is enable https flag.
	EnableHTTPS        bool
	defaultEnableHTTPS = false

	// TrustedSubnet is trusted subnet.
	// 192.168.1.0/24 (CIDR)
	TrustedSubnet        string
	defaultTrustedSubnet = ""

	// Config is config filename.
	Config string
)

// ParseFlags reads the startup flags and environment variables.
//
// If there is an environment variable, then it is used.
// If there is no environment variable, but there is a flag, then a flag is used.
// If there is json config flag, then it is used.
// If there is nothing, the default flag value is used.
func ParseFlags() {
	// getting basic flags
	flag.StringVar(&ServerAddress, "a", "", "address and port to run http server")
	flag.StringVar(&GRPCServerAddress, "g", "", "address and port to run grpc server")
	flag.StringVar(&BaseURL, "b", "", "address and port for include in shortURLs")
	flag.StringVar(&FileStoragePath, "f", "", "path to the data storage file")
	flag.StringVar(&DatabaseDSN, "d", "", "database connection string")
	flag.BoolVar(&EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&TrustedSubnet, "t", "", "trusted subnet")
	// getting additional flags
	flag.StringVar(&SecretKey, "k", "", "the secret key for user tokens")
	flag.StringVar(&SecureFilePath, "sec", "", "path to the secure data file")
	flag.StringVar(&LogLevel, "l", "", "logging level")
	// getting config.json file flag
	flag.StringVar(&Config, "config", "", "config filename")
	flag.StringVar(&Config, "c", "", "config filename")
	// parsing flags
	flag.Parse()

	// replacing from env
	envflags.TryUseEnvString(&ServerAddress, "SERVER_ADDRESS")
	envflags.TryUseEnvString(&GRPCServerAddress, "GRPC_SERVER_ADDRESS")
	envflags.TryUseEnvString(&BaseURL, "BASE_URL")
	envflags.TryUseEnvString(&FileStoragePath, "FILE_STORAGE_PATH")
	envflags.TryUseEnvString(&DatabaseDSN, "DATABASE_DSN")
	envflags.TryUseEnvBool(&EnableHTTPS, "ENABLE_HTTPS")
	envflags.TryUseEnvString(&Config, "CONFIG")
	envflags.TryUseEnvString(&TrustedSubnet, "TRUSTED_SUBNET")
	// additional env
	envflags.TryUseEnvString(&SecretKey, "SECRET_KEY")
	envflags.TryUseEnvString(&SecureFilePath, "SECURE_FILE_PATH")
	envflags.TryUseEnvString(&LogLevel, "LOG_LEVEL")

	// using config file or set default values
	if Config != "" {
		conf, er := getJSONConfig(Config)
		if er != nil {
			log.Panicf("error getting json config %s, error: %s", Config, er.Error())
		}
		// checking all config variables
		envflags.TryConfigStringFlag(&ServerAddress, conf.ServerAddress)
		envflags.TryConfigStringFlag(&GRPCServerAddress, conf.GRPCServerAddress)
		envflags.TryConfigStringFlag(&BaseURL, conf.BaseURL)
		envflags.TryConfigStringFlag(&FileStoragePath, conf.FileStoragePath)
		envflags.TryConfigStringFlag(&DatabaseDSN, conf.DatabaseDSN)
		envflags.TryConfigBoolFlag(&EnableHTTPS, conf.EnableHTTPS)
		envflags.TryConfigStringFlag(&TrustedSubnet, conf.TrustedSubnet)
		// additional variables
		envflags.TryConfigStringFlag(&SecretKey, conf.SecretKey)
		envflags.TryConfigStringFlag(&SecureFilePath, conf.SecureFilePath)
		envflags.TryConfigStringFlag(&LogLevel, conf.LogLevel)
	}

	// setting defaults
	// checking all config variables
	envflags.TryDefaultStringFlag(&ServerAddress, defaultServerAddress)
	envflags.TryDefaultStringFlag(&GRPCServerAddress, defaultGRPCServerAddress)
	envflags.TryDefaultStringFlag(&BaseURL, defaultBaseURL)
	envflags.TryDefaultStringFlag(&FileStoragePath, defaultFileStoragePath)
	envflags.TryDefaultStringFlag(&DatabaseDSN, defaultDatabaseDSN)
	envflags.TryDefaultBoolFlag(&EnableHTTPS, defaultEnableHTTPS)
	envflags.TryDefaultStringFlag(&TrustedSubnet, defaultTrustedSubnet)
	// additional variables
	envflags.TryDefaultStringFlag(&SecretKey, defaultSecretKey)
	envflags.TryDefaultStringFlag(&SecureFilePath, defaultSecureFilePath)
	envflags.TryDefaultStringFlag(&LogLevel, defaultLogLevel)

}
