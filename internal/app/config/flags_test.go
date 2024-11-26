package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {

	t.Run("replacing by env", func(t *testing.T) {
		_ = os.Setenv("SERVER_ADDRESS", "localhost:8888")
		_ = os.Setenv("BASE_URL", "localhost:9999")
		_ = os.Setenv("FILE_STORAGE_PATH", "./storage.db")
		_ = os.Setenv("DATABASE_DSN", "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable")
		_ = os.Setenv("SECRET_KEY", " ")
		_ = os.Setenv("SECURE_FILE_PATH", "./sec")
		_ = os.Setenv("LOG_LEVEL", "debug")
		ParseFlags()

		assert.Equal(t, "localhost:8888", ServerAddress)
		assert.Equal(t, "localhost:9999", BaseURL)
		assert.Equal(t, "./storage.db", FileStoragePath)
		assert.Equal(t, "host=127.0.0.1 user=shortener password=pass dbname=shortener sslmode=disable", DatabaseDSN)
		assert.Equal(t, " ", SecretKey)
		assert.Equal(t, "./sec", SecureFilePath)
		assert.Equal(t, "debug", LogLevel)
	})

}
