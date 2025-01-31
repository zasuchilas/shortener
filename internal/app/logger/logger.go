// Package logger sets up logging in the service.
package logger

import (
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/config"
)

// Variables
var (
	// Log is th global variable for logging access from anywhere in the application.
	Log = zap.NewNop()
)

// Initialize initializes logging.
func Initialize(level string) error {
	// parsing level
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// setting the configuration
	cfg := zap.NewDevelopmentConfig()
	//cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	// creating a logger based on the configuration
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	// setting the logger globally
	Log = zl

	return nil
}

// ServiceInfo displays general information about the service.
func ServiceInfo(appVersion string) {
	// get app module name
	moduleName := "-"
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		Log.Error("Failed to read build info")
	} else {
		moduleName = buildInfo.Main.Path
	}

	// write data to the log
	Log.Info(
		"SHORTENER (URL shortening service)",
		zap.String("name", moduleName),
		zap.String("version", appVersion),
	)

	//
	ConfigInfo()
}

// ConfigInfo logs config values.
func ConfigInfo() {
	Log.Info("Current app config",
		zap.String("ServerAddress", config.ServerAddress),
		zap.String("BaseURL", config.BaseURL),
		zap.String("FileStoragePath", config.FileStoragePath),
		zap.String("DatabaseDSN", config.DatabaseDSN),
	)
}
