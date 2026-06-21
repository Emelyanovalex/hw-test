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

	s := sender.New(consumer, logg)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	return s.Run(ctx)
}
