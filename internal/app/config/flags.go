package config

import (
	"flag"
	"os"
)

var (
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	SecretKey       string
	SecureFilePath  string
)

func ParseFlags() {
	// using flags (and set default values)
	flag.StringVar(&ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&BaseURL, "b", "localhost:8080", "address and port for include in shortURLs")
	flag.StringVar(&FileStoragePath, "f", "", "path to the data storage file")
	flag.StringVar(&DatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&SecretKey, "k", "supersecretkey", "the secret key for user tokens")
	flag.StringVar(&SecureFilePath, "s", "./secure.db", "path to the secure data file")
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
}

// TODO: validate flags
//netAddrRegex = regexp.MustCompile(`^[a-zA-Z0-9ЁёА-я.-]+\.[a-zA-ZЁёА-я0-9]{2,}:[0-9]{3,4}$`)
//func validateNetAddr(val string) error {
//	if !netAddrRegex.MatchString(val) {
//		return errors.New("")
//	}
//	return nil
//}
