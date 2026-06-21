package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all scheduler service configuration.
type Config struct {
	Logger    LoggerConf    `mapstructure:"logger"`
	Storage   StorageConf   `mapstructure:"storage"`
	Database  DatabaseConf  `mapstructure:"database"`
	RabbitMQ  RabbitMQConf  `mapstructure:"rabbitmq"`
	Scheduler SchedulerConf `mapstructure:"scheduler"`
}

// LoggerConf configures the logger.
type LoggerConf struct {
	Level string `mapstructure:"level"`
}

// StorageConf selects the storage backend ("memory" or "sql").
type StorageConf struct {
	Kind string `mapstructure:"kind"`
}

// DatabaseConf is used when Storage.Kind == "sql".
type DatabaseConf struct {
	DSN string `mapstructure:"dsn"`
}

// RabbitMQConf holds the AMQP connection string.
type RabbitMQConf struct {
	DSN string `mapstructure:"dsn"`
}

// SchedulerConf controls how often the scheduler runs.
type SchedulerConf struct {
	Interval time.Duration `mapstructure:"interval"`
}

const defaultSchedulerInterval = 10 * time.Second

// LoadConfig reads configuration from path and applies SCHEDULER_ env overrides.
func LoadConfig(path string) (Config, error) {
	v := viper.New()

	v.SetDefault("logger.level", "info")
	v.SetDefault("storage.kind", "sql")
	v.SetDefault("database.dsn", "")
	v.SetDefault("rabbitmq.dsn", "amqp://rabbit:password@localhost:5672/")
	v.SetDefault("scheduler.interval", defaultSchedulerInterval)

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config %q: %w", path, err)
		}
	}

	v.SetEnvPrefix("SCHEDULER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
