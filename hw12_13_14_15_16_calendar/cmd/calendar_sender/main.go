package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Emelyanovalex/hw12_calendar/internal/logger"
	"github.com/Emelyanovalex/hw12_calendar/internal/queue/rabbitmq"
	"github.com/Emelyanovalex/hw12_calendar/internal/sender"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/sender_config.yaml", "Path to configuration file")
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

	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQ.DSN)
	if err != nil {
		return fmt.Errorf("init rabbitmq consumer: %w", err)
	}
	defer func() { _ = consumer.Close() }()

	var sw sender.StatusWriter
	if cfg.Database.DSN != "" {
		dbLog, dbErr := sender.NewDBLog(cfg.Database.DSN)
		if dbErr != nil {
			return fmt.Errorf("init db log: %w", dbErr)
		}
		defer func() { _ = dbLog.Close() }()
		sw = dbLog
	}

	s := sender.New(consumer, logg, sw)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	return s.Run(ctx)
}
