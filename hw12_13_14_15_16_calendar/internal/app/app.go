package app

import (
	"context"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
)

// Logger is the minimal logger surface required by the app.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error)
}

type App struct {
	logger  Logger
	storage Storage
}

func New(logger Logger, storage Storage) *App {
	return &App{logger: logger, storage: storage}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.CreateEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	return a.storage.UpdateEvent(ctx, id, event)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	return a.storage.DeleteEvent(ctx, id)
}

func (a *App) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForDay(ctx, date)
}

func (a *App) ListEventsForWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForWeek(ctx, weekStart)
}

func (a *App) ListEventsForMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForMonth(ctx, monthStart)
}
