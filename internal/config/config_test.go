package config

import "testing"

func TestConfigDefaults(t *testing.T) {
	cfg := Load()
	if cfg.DatabasePath == "" || cfg.HTTPAddress == "" || cfg.PageSize < 1 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}
