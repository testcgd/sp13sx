package config

import "fmt"

func Validate(cfg Config) error {
	if cfg.Defaults.Backend == "" {
		return fmt.Errorf("defaults.backend is required")
	}
	if len(cfg.Backends) == 0 {
		return fmt.Errorf("at least one backend must be configured")
	}
	if _, ok := cfg.Backends[cfg.Defaults.Backend]; !ok {
		return fmt.Errorf("defaults.backend %q not found in backends", cfg.Defaults.Backend)
	}
	return nil
}
