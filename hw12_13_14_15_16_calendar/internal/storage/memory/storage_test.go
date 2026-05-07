package memorystorage

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func newEvent(id, userID string, start time.Time, dur time.Duration) storage.Event {
	return storage.Event{
		ID:        id,
		Title:     "title-" + id,
		StartTime: start,
		Duration:  dur,
		UserID:    userID,
	}
}

func TestStorage_CRUD(t *testing.T) {
	ctx := context.Background()
	s := New()
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	t.Run("create and list", func(t *testing.T) {
		e := newEvent("1", "u1", now, time.Hour)
		require.NoError(t, s.CreateEvent(ctx, e))

		events, err := s.ListEventsForDay(ctx, now)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, "1", events[0].ID)
	})

	t.Run("update existing event", func(t *testing.T) {
		e := newEvent("1", "u1", now, 2*time.Hour)
		e.Title = "renamed"
		require.NoError(t, s.UpdateEvent(ctx, "1", e))

		events, _ := s.ListEventsForDay(ctx, now)
		require.Equal(t, "renamed", events[0].Title)
		require.Equal(t, 2*time.Hour, events[0].Duration)
	})

	t.Run("update missing event", func(t *testing.T) {
		err := s.UpdateEvent(ctx, "missing", newEvent("missing", "u1", now, time.Hour))
		require.ErrorIs(t, err, storage.ErrEventNotFound)
	})

	t.Run("delete existing", func(t *testing.T) {
		require.NoError(t, s.DeleteEvent(ctx, "1"))
		events, _ := s.ListEventsForDay(ctx, now)
		require.Empty(t, events)
	})

	t.Run("delete missing", func(t *testing.T) {
		err := s.DeleteEvent(ctx, "missing")
		require.ErrorIs(t, err, storage.ErrEventNotFound)
	})
}

func TestStorage_DateBusy(t *testing.T) {
	ctx := context.Background()
	s := New()
	now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	require.NoError(t, s.CreateEvent(ctx, newEvent("1", "u1", now, time.Hour)))

	t.Run("duplicate id", func(t *testing.T) {
		err := s.CreateEvent(ctx, newEvent("1", "u2", now.Add(5*time.Hour), time.Hour))
		require.ErrorIs(t, err, storage.ErrDateBusy)
	})

	t.Run("overlapping for same user", func(t *testing.T) {
		err := s.CreateEvent(ctx, newEvent("2", "u1", now.Add(30*time.Minute), time.Hour))
		require.ErrorIs(t, err, storage.ErrDateBusy)
	})

	t.Run("non-overlapping for same user", func(t *testing.T) {
		err := s.CreateEvent(ctx, newEvent("3", "u1", now.Add(2*time.Hour), time.Hour))
		require.NoError(t, err)
	})

	t.Run("overlapping for different user is allowed", func(t *testing.T) {
		err := s.CreateEvent(ctx, newEvent("4", "u2", now, time.Hour))
		require.NoError(t, err)
	})

	t.Run("update creates overlap", func(t *testing.T) {
		require.NoError(t, s.CreateEvent(ctx, newEvent("5", "u3", now, time.Hour)))
		require.NoError(t, s.CreateEvent(ctx, newEvent("6", "u3", now.Add(2*time.Hour), time.Hour)))

		updated := newEvent("6", "u3", now.Add(30*time.Minute), time.Hour)
		err := s.UpdateEvent(ctx, "6", updated)
		require.ErrorIs(t, err, storage.ErrDateBusy)
	})
}

func TestStorage_ListRanges(t *testing.T) {
	ctx := context.Background()
	s := New()
	base := time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC) // Monday

	require.NoError(t, s.CreateEvent(ctx, newEvent("today", "u1", base, time.Hour)))
	require.NoError(t, s.CreateEvent(ctx, newEvent("tomorrow", "u1", base.AddDate(0, 0, 1), time.Hour)))
	require.NoError(t, s.CreateEvent(ctx, newEvent("nextweek", "u1", base.AddDate(0, 0, 8), time.Hour)))
	require.NoError(t, s.CreateEvent(ctx, newEvent("nextmonth", "u1", base.AddDate(0, 1, 1), time.Hour)))

	day, err := s.ListEventsForDay(ctx, base)
	require.NoError(t, err)
	require.Len(t, day, 1)

	week, err := s.ListEventsForWeek(ctx, base)
	require.NoError(t, err)
	require.Len(t, week, 2)

	month, err := s.ListEventsForMonth(ctx, base)
	require.NoError(t, err)
	require.Len(t, month, 3)
}

func TestStorage_Concurrent(t *testing.T) {
	ctx := context.Background()
	s := New()
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	const workers = 50
	const perWorker = 50

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(w int) {
			defer wg.Done()
			user := "u" + strconv.Itoa(w)
			for i := 0; i < perWorker; i++ {
				id := user + "-" + strconv.Itoa(i)
				start := base.Add(time.Duration(i) * 2 * time.Hour)
				err := s.CreateEvent(ctx, newEvent(id, user, start, time.Hour))
				if err != nil && !errors.Is(err, storage.ErrDateBusy) {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}
		}(w)
	}
	wg.Wait()

	month, err := s.ListEventsForMonth(ctx, base)
	require.NoError(t, err)
	require.Equal(t, workers*perWorker, len(month))
}
