package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
)

const daysInWeek = 7

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) CreateEvent(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[event.ID]; ok {
		return storage.ErrDateBusy
	}
	if s.hasOverlap(event, "") {
		return storage.ErrDateBusy
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, id string, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	if s.hasOverlap(event, id) {
		return storage.ErrDateBusy
	}

	event.ID = id
	s.events[id] = event
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	delete(s.events, id)
	return nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	start := truncateToDay(date)
	return s.listInRange(ctx, start, start.AddDate(0, 0, 1))
}

func (s *Storage) ListEventsForWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error) {
	start := truncateToDay(weekStart)
	return s.listInRange(ctx, start, start.AddDate(0, 0, daysInWeek))
}

func (s *Storage) ListEventsForMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error) {
	start := truncateToDay(monthStart)
	return s.listInRange(ctx, start, start.AddDate(0, 1, 0))
}

func (s *Storage) listInRange(_ context.Context, from, to time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]storage.Event, 0)
	for _, e := range s.events {
		if !e.StartTime.Before(from) && e.StartTime.Before(to) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Storage) hasOverlap(candidate storage.Event, ignoreID string) bool {
	candStart := candidate.StartTime
	candEnd := candidate.StartTime.Add(candidate.Duration)
	for _, e := range s.events {
		if e.ID == ignoreID || e.UserID != candidate.UserID {
			continue
		}
		eStart := e.StartTime
		eEnd := e.StartTime.Add(e.Duration)
		if candStart.Before(eEnd) && eStart.Before(candEnd) {
			return true
		}
	}
	return false
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
