package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/logger"
	"github.com/Emelyanovalex/hw12_calendar/internal/queue/rabbitmq"
	"github.com/Emelyanovalex/hw12_calendar/internal/scheduler"
	memorystorage "github.com/Emelyanovalex/hw12_calendar/internal/storage/memory"
	sqlstorage "github.com/Emelyanovalex/hw12_calendar/internal/storage/sql"
)

const (
	dbConnectTimeout = 5 * time.Second
	dbCloseTimeout   = 3 * time.Second
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/scheduler_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logg := logger.New(cfg.Logger.Level)
	defer func() { _ = logg.Sync() }()

	stor, cleanup, err := buildStorage(cfg)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	defer cleanup()

	producer, err := rabbitmq.NewProducer(cfg.RabbitMQ.DSN)
	if err != nil {
		return fmt.Errorf("init rabbitmq producer: %w", err)
	}
	defer func() { _ = producer.Close() }()

	sched := scheduler.New(stor, producer, logg, cfg.Scheduler.Interval)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	return sched.Run(ctx)
}

func buildStorage(cfg Config) (scheduler.Storage, func(), error) {
	switch cfg.Storage.Kind {
	case "", "memory":
		return memorystorage.New(), func() {}, nil
	case "sql":
		s := sqlstorage.New(cfg.Database.DSN)
		ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
		defer cancel()
		if err := s.Connect(ctx); err != nil {
			return nil, nil, err
		}
		cleanup := func() {
			closeCtx, closeCancel := context.WithTimeout(context.Background(), dbCloseTimeout)
			defer closeCancel()
			_ = s.Close(closeCtx)
		}
		return s, cleanup, nil
	default:
		return nil, nil, fmt.Errorf("unknown storage kind %q", cfg.Storage.Kind)
	}
}
