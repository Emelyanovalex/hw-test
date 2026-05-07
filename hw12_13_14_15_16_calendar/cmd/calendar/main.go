package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/app"
	"github.com/Emelyanovalex/hw12_calendar/internal/logger"
	internalhttp "github.com/Emelyanovalex/hw12_calendar/internal/server/http"
	memorystorage "github.com/Emelyanovalex/hw12_calendar/internal/storage/memory"
	sqlstorage "github.com/Emelyanovalex/hw12_calendar/internal/storage/sql"
)

const (
	dbConnectTimeout       = 5 * time.Second
	dbCloseTimeout         = 3 * time.Second
	defaultShutdownTimeout = 3 * time.Second
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

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

	calendar := app.New(logg, stor)

	server := internalhttp.NewServer(logg, calendar, internalhttp.Config{
		Host:            cfg.HTTP.Host,
		Port:            cfg.HTTP.Port,
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	})

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		shutdownTimeout := cfg.HTTP.ShutdownTimeout
		if shutdownTimeout <= 0 {
			shutdownTimeout = defaultShutdownTimeout
		}
		stopCtx, stopCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer stopCancel()

		if stopErr := server.Stop(stopCtx); stopErr != nil {
			logg.Error("failed to stop http server: " + stopErr.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

func buildStorage(cfg Config) (app.Storage, func(), error) {
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
