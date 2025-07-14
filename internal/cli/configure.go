package cli

import (
	"fmt"

	"github.com/spf13/viper"
)

// Configure handles the configuration process
func Configure(v *viper.Viper, configFile, hostname, username, password string) error {
	// Load current configuration
	cfg, err := load(v, configFile)
	if err != nil {
		return err
	}

	// Update with new values from flags
	cfg.update(hostname, username, password)

	// Prompt for missing values
	if err := promptForValues(cfg); err != nil {
		return err
	}

	// Save configuration
	if err := cfg.save(v); err != nil {
		return err
	}

	// Print summary
	cfg.printSummary(v)
	return nil
}

// ShowConfiguration displays the current configuration
func ShowConfiguration(v *viper.Viper, configFile string) error {
	// Load current configuration
	cfg, err := load(v, configFile)
	if err != nil {
		return err
	}

	// Print current values
	cfg.printSummary(v)
	return nil
}

// promptForValues prompts the user for missing configuration values
func promptForValues(cfg *Config) error {
	fields := []struct {
		displayName string
		maskValue   bool
		getter      func() string
		setter      func(string)
	}{
		{
			displayName: "Hostname/IP address",
			maskValue:   false,
			getter:      func() string { return cfg.Hostname },
			setter:      func(s string) { cfg.Hostname = s },
		},
		{
			displayName: "Username",
			maskValue:   false,
			getter:      func() string { return cfg.Username },
			setter:      func(s string) { cfg.Username = s },
		},
		{
			displayName: "Password",
			maskValue:   true,
			getter:      func() string { return cfg.Password },
			setter:      func(s string) { cfg.Password = s },
		},
	}

	for _, field := range fields {
		if err := promptForField(field.displayName, field.getter(), field.maskValue, field.setter); err != nil {
			return err
		}
	}
	return nil
}

// promptForField prompts for a single field value
func promptForField(displayName, currentValue string, maskValue bool, setter func(string)) error {
	var newValue string

	if currentValue != "" {
		if maskValue {
			fmt.Printf("%s [***]: ", displayName)
		} else {
			fmt.Printf("%s [%s]: ", displayName, currentValue)
		}
	} else {
		fmt.Printf("%s: ", displayName)
	}

	_, err := fmt.Scanln(&newValue)
	if err != nil {
		// Handle the case where user just presses Enter (empty input)
		if err.Error() == "unexpected newline" {
			newValue = currentValue
		} else {
			return fmt.Errorf("error reading input: %w", err)
		}
	}

	if newValue == "" {
		newValue = currentValue
	}

	setter(newValue)
	return nil
}
