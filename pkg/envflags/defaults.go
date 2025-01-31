package envflags

// TryDefaultStringFlag tries to update the flag values
// from the default value (string type).
func TryDefaultStringFlag(flagValue *string, defaultValue string) {
	// value already set
	if *flagValue != "" {
		return
	}

	// setting default value
	*flagValue = defaultValue
}

// TryDefaultBoolFlag tries to update the flag values
// from the default value (bool type).
func TryDefaultBoolFlag(flagValue *bool, defaultValue bool) {
	// value already set
	// it means flag.BoolVar or ENV or config.json already changed value
	// (because default for flag.BoolVar set false)
	if *flagValue {
		return
	}

	// if code default is true then flag must be true
	*flagValue = defaultValue
}
