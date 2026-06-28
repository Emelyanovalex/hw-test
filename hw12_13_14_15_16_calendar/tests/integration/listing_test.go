package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestListEventsForDay(t *testing.T) {
	userID := randomID()
	day := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 10)

	id1 := randomID()
	body1 := buildEventJSON(id1, "Day Event 1", day.Add(9*time.Hour), time.Hour, userID, 0)
	resp1, err := apiRequest(http.MethodPost, "/events", body1)
	require.NoError(t, err)
	_ = resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	id2 := randomID()
	body2 := buildEventJSON(id2, "Day Event 2", day.Add(14*time.Hour), time.Hour, userID, 0)
	resp2, err := apiRequest(http.MethodPost, "/events", body2)
	require.NoError(t, err)
	_ = resp2.Body.Close()
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	dateStr := day.Format("2006-01-02")
	resp, err := apiRequest(http.MethodGet, fmt.Sprintf("/events/day?date=%s", dateStr), "")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	events, ok := result["events"].([]any)
	require.True(t, ok)

	ids := make(map[string]bool)
	for _, e := range events {
		ev, _ := e.(map[string]any)
		ids[ev["id"].(string)] = true
	}
	require.True(t, ids[id1], "event 1 should be in day listing")
	require.True(t, ids[id2], "event 2 should be in day listing")
}

func TestListEventsForWeek(t *testing.T) {
	userID := randomID()
	weekStart := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 20)

	id1 := randomID()
	body1 := buildEventJSON(id1, "Week Event 1", weekStart.Add(time.Hour), time.Hour, userID, 0)
	resp1, err := apiRequest(http.MethodPost, "/events", body1)
	require.NoError(t, err)
	_ = resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	id2 := randomID()
	body2 := buildEventJSON(id2, "Week Event 2", weekStart.AddDate(0, 0, 3).Add(time.Hour), time.Hour, userID, 0)
	resp2, err := apiRequest(http.MethodPost, "/events", body2)
	require.NoError(t, err)
	_ = resp2.Body.Close()
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	idOther := randomID()
	bodyOther := buildEventJSON(idOther, "Outside Week", weekStart.AddDate(0, 0, 10).Add(time.Hour), time.Hour, userID, 0)
	respOther, err := apiRequest(http.MethodPost, "/events", bodyOther)
	require.NoError(t, err)
	_ = respOther.Body.Close()
	require.Equal(t, http.StatusCreated, respOther.StatusCode)

	dateStr := weekStart.Format("2006-01-02")
	resp, err := apiRequest(http.MethodGet, fmt.Sprintf("/events/week?date=%s", dateStr), "")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	events, ok := result["events"].([]any)
	require.True(t, ok)

	ids := make(map[string]bool)
	for _, e := range events {
		ev, _ := e.(map[string]any)
		ids[ev["id"].(string)] = true
	}
	require.True(t, ids[id1], "event 1 should be in week listing")
	require.True(t, ids[id2], "event 2 should be in week listing")
	require.False(t, ids[idOther], "event outside week should not appear")
}

func TestListEventsForMonth(t *testing.T) {
	userID := randomID()
	monthStart := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 2, 0)
	monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, time.UTC)

	id1 := randomID()
	body1 := buildEventJSON(id1, "Month Event 1", monthStart.Add(time.Hour), time.Hour, userID, 0)
	resp1, err := apiRequest(http.MethodPost, "/events", body1)
	require.NoError(t, err)
	_ = resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	id2 := randomID()
	body2 := buildEventJSON(id2, "Month Event 2", monthStart.AddDate(0, 0, 15).Add(time.Hour), time.Hour, userID, 0)
	resp2, err := apiRequest(http.MethodPost, "/events", body2)
	require.NoError(t, err)
	_ = resp2.Body.Close()
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	idNext := randomID()
	bodyNext := buildEventJSON(idNext, "Next Month", monthStart.AddDate(0, 1, 1).Add(time.Hour), time.Hour, userID, 0)
	respNext, err := apiRequest(http.MethodPost, "/events", bodyNext)
	require.NoError(t, err)
	_ = respNext.Body.Close()
	require.Equal(t, http.StatusCreated, respNext.StatusCode)

	dateStr := monthStart.Format("2006-01-02")
	resp, err := apiRequest(http.MethodGet, fmt.Sprintf("/events/month?date=%s", dateStr), "")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	events, ok := result["events"].([]any)
	require.True(t, ok)

	ids := make(map[string]bool)
	for _, e := range events {
		ev, _ := e.(map[string]any)
		ids[ev["id"].(string)] = true
	}
	require.True(t, ids[id1], "event 1 should be in month listing")
	require.True(t, ids[id2], "event 2 should be in month listing")
	require.False(t, ids[idNext], "next month event should not appear")
}
