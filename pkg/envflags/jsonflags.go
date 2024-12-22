package envflags

// TryConfigStringFlag tries to update the flag values
// from the json config (string type).
func TryConfigStringFlag(flagValue *string, configValue string) {
	// value already set
	if *flagValue != "" {
		return
	}

	// trying setting json config value
	if configValue != "" {
		*flagValue = configValue
	}
}

// TryConfigBoolFlag tries to update the flag values
// from the json config (bool type).
func TryConfigBoolFlag(flagValue *bool, configValue bool) {
	// value already set
	// it means flag.BoolVar or ENV already changed value (default for flag.BoolVar is false)
	if *flagValue {
		return
	}

	// trying setting json config value
	// default/untouched false may be resetting by json config
	if configValue {
		*flagValue = configValue
	}
}
