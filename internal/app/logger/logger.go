package logger

import (
	"log"
	"runtime/debug"
	"time"
)

func ServiceInfo(appVersion string) {
	// get app module name
	buildInfo, ok := debug.ReadBuildInfo()
	var moduleName string
	if !ok {
		log.Printf("Failed to read build info")
		moduleName = "-"
	} else {
		moduleName = buildInfo.Main.Path
	}

	// write data to the log
	log.Printf("*****************************************************************")
	log.Printf("SHORTENER (URL shortening service)")
	log.Printf("%s v%s", moduleName, appVersion)
	log.Printf("Started at %s", time.Now().UTC().Format(time.RFC1123))
	log.Printf("*****************************************************************")

}
