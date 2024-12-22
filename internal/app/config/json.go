package config

import (
	"encoding/json"
	"os"
)

type jsonConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`

	SecretKey      string `json:"secret_key"`
	SecureFilePath string `json:"secure_file_path"`
	LogLevel       string `json:"log_level"`
}

func getJsonConfig(filename string) (*jsonConfig, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var jc jsonConfig
	err = json.Unmarshal(file, &jc)
	if err != nil {
		return nil, err
	}

	return &jc, nil
}
