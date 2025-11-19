package config

import (
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application.
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Client     ClientConfig     `yaml:"client"`
	Queue      QueueConfig      `yaml:"queue"`
	Logging    LoggingConfig    `yaml:"logging"`
	CIDRules   CIDRules         `yaml:"cidrules"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	UI         UIConfig         `yaml:"ui"`
}

// ServerConfig holds server-specific configuration.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// ClientConfig holds client-specific configuration.
type ClientConfig struct {
	Host             string        `yaml:"host"`
	Port             string        `yaml:"port"`
	ReconnectInitial time.Duration `yaml:"reconnectinitial"`
	ReconnectMax     time.Duration `yaml:"reconnectmax"`
}

// QueueConfig holds queue-specific configuration.
type QueueConfig struct {
	BufferSize int `yaml:"buffersize"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxsize"`
	MaxBackups int    `yaml:"maxbackups"`
	MaxAge     int    `yaml:"maxage"`
	Compress   bool   `yaml:"compress"`
	Level      string `yaml:"level"`
}

// CIDRules holds the specific rules for CID message processing.
type CIDRules struct {
	RequiredPrefix string            `yaml:"requiredprefix"`
	ValidLength    int               `yaml:"validlength"`
	TestCodeMap    map[string]string `yaml:"testcodemap"`
	AccNumOffset   int               `yaml:"accnumoffset"`
	AccNumAdd      int               `yaml:"accnumadd"`
}

// MonitoringConfig holds configuration for UI monitoring.
type MonitoringConfig struct {
	PPKTimeout time.Duration `yaml:"ppktimeout"`
}

// UIConfig holds UI-specific configuration.
type UIConfig struct {
	StartMinimized bool `yaml:"startminimized"` // Start application minimized to tray
	MinimizeToTray bool `yaml:"minimizetotray"` // Minimize to tray instead of taskbar
	CloseToTray    bool `yaml:"closetotray"`    // Close button minimizes to tray instead of exiting
}

// defaultConfig returns a new Config with default values.
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "20005",
		},
		Client: ClientConfig{
			Host:             "10.32.1.49",
			Port:             "20004",
			ReconnectInitial: 1 * time.Second,
			ReconnectMax:     60 * time.Second,
		},
		Queue: QueueConfig{
			BufferSize: 100,
		},
		Logging: LoggingConfig{
			Filename:   "app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     28,
			Compress:   true,
			Level:      "INFO",
		},
		CIDRules: CIDRules{
			RequiredPrefix: "5",
			ValidLength:    21,
			TestCodeMap:    map[string]string{"E603": "E602"},
			AccNumOffset:   2100,
			AccNumAdd:      2100,
		},
		Monitoring: MonitoringConfig{
			PPKTimeout: 15 * time.Minute,
		},
		UI: UIConfig{
			StartMinimized: false,
			MinimizeToTray: false,
			CloseToTray:    false,
		},
	}
}

// New loads the configuration from the default path ("config.yaml").
// If the file does not exist, it creates a default one.
// It panics if any other error occurs, as config is critical.
func New() *Config {
	path := "config.yaml"
	cfg, err := load(path)
	if err == nil {
		return cfg
	}

	if !os.IsNotExist(err) {
		slog.Error("Failed to load configuration", "path", path, "error", err)
		panic(err)
	}

	slog.Warn("Configuration file not found, creating a default one.", "path", path)
	defaultCfg := defaultConfig()

	data, err := yaml.Marshal(defaultCfg)
	if err != nil {
		slog.Error("Failed to marshal default configuration", "error", err)
		panic(err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		slog.Error("Failed to write default configuration file", "path", path, "error", err)
		panic(err)
	}

	slog.Info("Default configuration file created successfully.", "path", path)
	return defaultCfg
}

// load reads the configuration file from the given path and unmarshals it.
func load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the current configuration to the file at the given path.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		slog.Error("Failed to marshal configuration", "error", err)
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		slog.Error("Failed to write configuration file", "path", path, "error", err)
		return err
	}

	slog.Info("Configuration saved successfully", "path", path)
	return nil
}
