package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all sender service configuration.
type Config struct {
	Logger   LoggerConf   `mapstructure:"logger"`
	RabbitMQ RabbitMQConf `mapstructure:"rabbitmq"`
}

// LoggerConf configures the logger.
type LoggerConf struct {
	Level string `mapstructure:"level"`
}

// RabbitMQConf holds the AMQP connection string.
type RabbitMQConf struct {
	DSN string `mapstructure:"dsn"`
}

// LoadConfig reads configuration from path and applies SENDER_ env overrides.
func LoadConfig(path string) (Config, error) {
	v := viper.New()

	v.SetDefault("logger.level", "info")
	v.SetDefault("rabbitmq.dsn", "amqp://rabbit:password@localhost:5672/")

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config %q: %w", path, err)
		}
	}

	v.SetEnvPrefix("SENDER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
