package logger

import (
	"go.uber.org/zap"
	"runtime/debug"
)

var (
	Log = zap.NewNop()
)

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

func ServiceInfo(appVersion string) {
	// get app module name
	buildInfo, ok := debug.ReadBuildInfo()
	var moduleName string
	if !ok {
		Log.Error("Failed to read build info")
		moduleName = "-"
	} else {
		moduleName = buildInfo.Main.Path
	}

	// write data to the log
	Log.Info(
		"SHORTENER (URL shortening service)",
		zap.String("name", moduleName),
		zap.String("version", appVersion),
	)
}
