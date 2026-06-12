package internalgrpc

import (
	"context"
	"testing"
	"time"

	api "github.com/Emelyanovalex/hw12_calendar/api"
	"github.com/Emelyanovalex/hw12_calendar/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockApp struct {
	events    map[string]storage.Event
	createErr error
	updateErr error
	deleteErr error
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
	return m.allEvents(), nil
}

func (m *mockApp) ListEventsForWeek(_ context.Context, _ time.Time) ([]storage.Event, error) {
	return m.allEvents(), nil
}

func (m *mockApp) ListEventsForMonth(_ context.Context, _ time.Time) ([]storage.Event, error) {
	return m.allEvents(), nil
}

func (m *mockApp) allEvents() []storage.Event {
	result := make([]storage.Event, 0, len(m.events))
	for _, e := range m.events {
		result = append(result, e)
	}
	return result
}

func newTestService(app Application) *calendarService {
	return &calendarService{app: app}
}

func TestCreateEvent(t *testing.T) {
	svc := newTestService(newMockApp())

	resp, err := svc.CreateEvent(context.Background(), &api.CreateEventRequest{
		Event: &api.Event{
			Id:        "1",
			Title:     "Meeting",
			StartTime: time.Now().Unix(),
			Duration:  int64(time.Hour),
			UserId:    "user1",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "1", resp.GetEvent().GetId())
	assert.Equal(t, "Meeting", resp.GetEvent().GetTitle())
}

func TestCreateEvent_DateBusy(t *testing.T) {
	app := newMockApp()
	app.createErr = storage.ErrDateBusy
	svc := newTestService(app)

	_, err := svc.CreateEvent(context.Background(), &api.CreateEventRequest{
		Event: &api.Event{Id: "1", Title: "Test", UserId: "user1"},
	})

	require.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestUpdateEvent_NotFound(t *testing.T) {
	svc := newTestService(newMockApp())

	_, err := svc.UpdateEvent(context.Background(), &api.UpdateEventRequest{
		Id:    "nonexistent",
		Event: &api.Event{Title: "Updated", UserId: "user1"},
	})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteEvent(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "To Delete", UserID: "user1"}
	svc := newTestService(app)

	_, err := svc.DeleteEvent(context.Background(), &api.DeleteEventRequest{Id: "1"})

	require.NoError(t, err)
	assert.Empty(t, app.events)
}

func TestDeleteEvent_NotFound(t *testing.T) {
	svc := newTestService(newMockApp())

	_, err := svc.DeleteEvent(context.Background(), &api.DeleteEventRequest{Id: "missing"})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListEventsForDay(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "Event", UserID: "user1"}
	svc := newTestService(app)

	resp, err := svc.ListEventsForDay(context.Background(), &api.ListEventsRequest{Date: time.Now().Unix()})

	require.NoError(t, err)
	assert.Len(t, resp.GetEvents(), 1)
}

func TestListEventsForWeek(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "Event", UserID: "user1"}
	svc := newTestService(app)

	resp, err := svc.ListEventsForWeek(context.Background(), &api.ListEventsRequest{Date: time.Now().Unix()})

	require.NoError(t, err)
	assert.Len(t, resp.GetEvents(), 1)
}

func TestListEventsForMonth(t *testing.T) {
	app := newMockApp()
	app.events["1"] = storage.Event{ID: "1", Title: "Event", UserID: "user1"}
	svc := newTestService(app)

	resp, err := svc.ListEventsForMonth(context.Background(), &api.ListEventsRequest{Date: time.Now().Unix()})

	require.NoError(t, err)
	assert.Len(t, resp.GetEvents(), 1)
}

func TestProtoEventConversion(t *testing.T) {
	now := time.Unix(time.Now().Unix(), 0).UTC()
	original := storage.Event{
		ID:           "42",
		Title:        "Test Event",
		StartTime:    now,
		Duration:     2 * time.Hour,
		Description:  "description",
		UserID:       "user42",
		NotifyBefore: 30 * time.Minute,
	}

	proto := eventToProto(original)
	assert.Equal(t, "42", proto.GetId())
	assert.Equal(t, "Test Event", proto.GetTitle())
	assert.Equal(t, now.Unix(), proto.GetStartTime())
	assert.Equal(t, int64(2*time.Hour), proto.GetDuration())

	back := protoToEvent(proto)
	assert.Equal(t, original, back)
}
