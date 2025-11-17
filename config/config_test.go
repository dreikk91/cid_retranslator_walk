package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNew_DefaultConfigCreation(t *testing.T) {
	// Ensure no config file exists
	configPath := "config.yaml"
	os.Remove(configPath)

	// Call New() to create a default config
	cfg := New()

	// Check if the file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("New() did not create the default config file at %s", configPath)
	}

	// Verify the created config matches the default
	defaultCfg := defaultConfig()
	if cfg.Server.Host != defaultCfg.Server.Host || cfg.Server.Port != defaultCfg.Server.Port {
		t.Errorf("New() created a config that does not match the default. Got %+v, want %+v", cfg.Server, defaultCfg.Server)
	}

	// Clean up the created file
	os.Remove(configPath)
}

func TestNew_LoadExistingConfig(t *testing.T) {
	// Create a custom config file
	configPath := "config.yaml"
	customCfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: "9999",
		},
	}
	data, _ := yaml.Marshal(customCfg)
	os.WriteFile(configPath, data, 0644)

	// Call New() to load the config
	loadedCfg := New()

	// Verify the loaded config matches the custom one
	if loadedCfg.Server.Host != customCfg.Server.Host || loadedCfg.Server.Port != customCfg.Server.Port {
		t.Errorf("New() did not load the existing config correctly. Got %+v, want %+v", loadedCfg.Server, customCfg.Server)
	}

	// Clean up the created file
	os.Remove(configPath)
}
