package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetJSONConfig(t *testing.T) {
	filename := "./mocks/config_mock.json"

	exp := jsonConfig{
		ServerAddress:     "localhost:33333",
		GRPCServerAddress: "localhost:33336",
		BaseURL:           ":33335",
		FileStoragePath:   "./storage_example.db",
		DatabaseDSN:       "host=127.0.0.1 user=shortener password=password dbname=shortener sslmode=disable",
		EnableHTTPS:       false,
		SecretKey:         "supersecretkey",
		SecureFilePath:    "./secure_example.db",
		LogLevel:          "debug",
	}

	res, err := getJSONConfig(filename)
	require.NoError(t, err)
	require.Equal(t, *res, exp)
}

func TestGetJSONConfigFilenameError(t *testing.T) {
	filename := "./mocks/config_mock_err.json"
	_, err := getJSONConfig(filename)
	t.Log(err)
	require.ErrorContains(t, err, "no such file or directory")
}

func TestGetJSONConfigUnmarshalErr(t *testing.T) {
	filename := "./mocks/config_mock_wrong.json"
	_, err := getJSONConfig(filename)
	t.Log(err)
	require.ErrorContains(t, err, "unexpected end of JSON")
}
