package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func randomID() string {
	return fmt.Sprintf("test-%x", rand.Int63())
}

func buildEventJSON(id, title string, start time.Time, duration time.Duration, userID string, notifyBefore time.Duration) string {
	return fmt.Sprintf(
		`{"id":%q,"title":%q,"start_time":%q,"duration":%d,"user_id":%q,"notify_before":%d}`,
		id, title, start.UTC().Format(time.RFC3339), int64(duration), userID, int64(notifyBefore),
	)
}

func TestCreateEvent(t *testing.T) {
	id := randomID()
	body := buildEventJSON(id, "Test Event", time.Now().UTC().Add(24*time.Hour), time.Hour, "user-create", 0)

	resp, err := apiRequest(http.MethodPost, "/events", body)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	event, ok := result["event"].(map[string]any)
	require.True(t, ok, "response should contain event object")
	require.Equal(t, id, event["id"])
	require.Equal(t, "Test Event", event["title"])
}

func TestCreateEventConflict(t *testing.T) {
	userID := randomID()
	start := time.Now().UTC().Add(48 * time.Hour)

	id1 := randomID()
	body1 := buildEventJSON(id1, "Event 1", start, time.Hour, userID, 0)
	resp1, err := apiRequest(http.MethodPost, "/events", body1)
	require.NoError(t, err)
	_ = resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	id2 := randomID()
	body2 := buildEventJSON(id2, "Event 2", start.Add(30*time.Minute), time.Hour, userID, 0)
	resp2, err := apiRequest(http.MethodPost, "/events", body2)
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()
	require.Equal(t, http.StatusConflict, resp2.StatusCode)
}

func TestUpdateEvent(t *testing.T) {
	id := randomID()
	start := time.Now().UTC().Add(72 * time.Hour)
	body := buildEventJSON(id, "Original Title", start, time.Hour, "user-update", 0)

	resp, err := apiRequest(http.MethodPost, "/events", body)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	updated := buildEventJSON(id, "Updated Title", start.Add(time.Hour), time.Hour, "user-update", 0)
	resp2, err := apiRequest(http.MethodPut, "/events/"+id, updated)
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	data, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))
	event, ok := result["event"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Updated Title", event["title"])
}

func TestDeleteEvent(t *testing.T) {
	id := randomID()
	body := buildEventJSON(id, "Delete Me", time.Now().UTC().Add(96*time.Hour), time.Hour, "user-delete", 0)

	resp, err := apiRequest(http.MethodPost, "/events", body)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp2, err := apiRequest(http.MethodDelete, "/events/"+id, "")
	require.NoError(t, err)
	_ = resp2.Body.Close()
	require.Equal(t, http.StatusNoContent, resp2.StatusCode)
}

func TestDeleteEventNotFound(t *testing.T) {
	resp, err := apiRequest(http.MethodDelete, "/events/nonexistent-id-xyz", "")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateEventNotFound(t *testing.T) {
	body := buildEventJSON("ghost", "Title", time.Now().UTC().Add(time.Hour), time.Hour, "user-1", 0)
	resp, err := apiRequest(http.MethodPut, "/events/ghost", body)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
