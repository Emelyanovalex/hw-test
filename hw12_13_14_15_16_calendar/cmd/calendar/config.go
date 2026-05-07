package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all calendar service configuration.
type Config struct {
	Logger   LoggerConf   `mapstructure:"logger"`
	HTTP     HTTPConf     `mapstructure:"http"`
	Storage  StorageConf  `mapstructure:"storage"`
	Database DatabaseConf `mapstructure:"database"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

type HTTPConf struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// StorageConf chooses which storage implementation to use.
// Allowed values: "memory", "sql".
type StorageConf struct {
	Kind string `mapstructure:"kind"`
}

// DatabaseConf is only used when Storage.Kind == "sql".
type DatabaseConf struct {
	DSN string `mapstructure:"dsn"`
}

// LoadConfig reads configuration from the given file path. The format is
// inferred from the file extension. Environment variables prefixed with
// CALENDAR_ override file values (e.g. CALENDAR_HTTP_PORT).
func LoadConfig(path string) (Config, error) {
	v := viper.New()

	v.SetDefault("logger.level", "info")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_timeout", "5s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.shutdown_timeout", "3s")
	v.SetDefault("storage.kind", "memory")
	v.SetDefault("database.dsn", "")

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config %q: %w", path, err)
		}
	}

	v.SetEnvPrefix("CALENDAR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
