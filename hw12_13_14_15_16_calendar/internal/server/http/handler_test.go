package internalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockApp implements Application for testing.
type mockApp struct {
	events    map[string]storage.Event
	createErr error
	updateErr error
	deleteErr error
	listErr   error
}

func newMockApp() *mockApp {
	return &mockApp{events: make(map[string]storage.Event)}
}

func (m *mockApp) CreateEvent(_ context.Context, event storage.Event) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.events[event.ID] = event
	return nil
}

func (m *mockApp) UpdateEvent(_ context.Context, id string, event storage.Event) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	event.ID = id
	m.events[id] = event
	return nil
}

func (m *mockApp) DeleteEvent(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	delete(m.events, id)
	return nil
}

func (m *mockApp) ListEventsForDay(_ context.Context, _ time.Time) ([]storage.Event, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.allEvents(), nil
}

func (m *mockApp) ListEventsForWeek(_ context.Context, _ time.Time) ([]storage.Event, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.allEvents(), nil
}

func (m *mockApp) ListEventsForMonth(_ context.Context, _ time.Time) ([]storage.Event, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.allEvents(), nil
}

func (m *mockApp) allEvents() []storage.Event {
	result := make([]storage.Event, 0, len(m.events))
	for _, e := range m.events {
		result = append(result, e)
	}
	return result
}

type noopLogger struct{}

func (n *noopLogger) Info(_ string)  {}
func (n *noopLogger) Error(_ string) {}

func newTestHandler(app *mockApp) *handler {
	return &handler{app: app, logger: &noopLogger{}}
}

const validEventBody = `{"id":"1","title":"Meeting","start_time":"2024-01-15T10:00:00Z","duration":3600000000000,"user_id":"user1"}`

func TestCreateEvent_Success(t *testing.T) {
	h := newTestHandler(newMockApp())
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(validEventBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.createEvent(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp eventResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "1", resp.Event.ID)
	assert.Equal(t, "Meeting", resp.Event.Title)
}

func TestCreateEvent_InvalidBody(t *testing.T) {
	h := newTestHandler(newMockApp())
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader("not-json"))
	rec := httptest.NewRecorder()

	h.createEvent(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateEvent_InvalidDate(t *testing.T) {
	h := newTestHandler(newMockApp())
	body := `{"id":"1","title":"T","start_time":"not-a-date","duration":0,"user_id":"u1"}`
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.createEvent(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateEvent_DateBusy(t *testing.T) {
	app := newMockApp()
	app.createErr = storage.ErrDateBusy
	h := newTestHandler(app)
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(validEventBody))
	rec := httptest.NewRecorder()

	h.createEvent(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestUpdateEvent_Success(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "Old", UserID: "user1", StartTime: time.Now()}
	h := newTestHandler(app)

	req := httptest.NewRequest(http.MethodPut, "/events/1", strings.NewReader(validEventBody))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.updateEvent(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateEvent_NotFound(t *testing.T) {
	h := newTestHandler(newMockApp())
	req := httptest.NewRequest(http.MethodPut, "/events/999", strings.NewReader(validEventBody))
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	h.updateEvent(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteEvent_Success(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "To Delete", UserID: "user1"}
	h := newTestHandler(app)

	req := httptest.NewRequest(http.MethodDelete, "/events/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.deleteEvent(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, app.events)
}

func TestDeleteEvent_NotFound(t *testing.T) {
	h := newTestHandler(newMockApp())
	req := httptest.NewRequest(http.MethodDelete, "/events/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	h.deleteEvent(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListEventsForDay_Success(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "Event", UserID: "u1", StartTime: time.Now()}
	h := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/day?date=2024-01-15", nil)
	rec := httptest.NewRecorder()

	h.listEventsForDay(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp eventsResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Len(t, resp.Events, 1)
}

func TestListEventsForDay_MissingDate(t *testing.T) {
	h := newTestHandler(newMockApp())
	req := httptest.NewRequest(http.MethodGet, "/events/day", nil)
	rec := httptest.NewRecorder()

	h.listEventsForDay(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListEventsForDay_InvalidDate(t *testing.T) {
	h := newTestHandler(newMockApp())
	req := httptest.NewRequest(http.MethodGet, "/events/day?date=not-a-date", nil)
	rec := httptest.NewRecorder()

	h.listEventsForDay(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListEventsForWeek_Success(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "Event", UserID: "u1", StartTime: time.Now()}
	h := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/week?date=2024-01-15", nil)
	rec := httptest.NewRecorder()

	h.listEventsForWeek(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListEventsForMonth_Success(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "Event", UserID: "u1", StartTime: time.Now()}
	h := newTestHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/events/month?date=2024-01-01", nil)
	rec := httptest.NewRecorder()

	h.listEventsForMonth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
