package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	notifyPollInterval = 2 * time.Second
	notifyTimeout      = 60 * time.Second
)

// TestNotificationSent verifies the full notification pipeline:
// event created → scheduler detects → RabbitMQ → sender → notifications_sent table.
func TestNotificationSent(t *testing.T) {
	eventID := randomID()

	// start_time is 10s in the future; notify_before is 30s.
	// start_time - notify_before = now - 20s, which is already past,
	// so the scheduler picks this up on its next tick.
	startTime := time.Now().UTC().Add(10 * time.Second)
	notifyBefore := 30 * time.Second

	body := buildEventJSON(eventID, "Notification Test", startTime, 2*time.Hour, "user-notify", notifyBefore)
	resp, err := apiRequest(http.MethodPost, "/events", body)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	deadline := time.Now().Add(notifyTimeout)
	for time.Now().Before(deadline) {
		var count int
		err := testDB.QueryRow(
			`SELECT COUNT(*) FROM notifications_sent WHERE event_id = $1`, eventID,
		).Scan(&count)
		require.NoError(t, err)
		if count > 0 {
			return
		}
		time.Sleep(notifyPollInterval)
	}

	t.Fatal("notification was not recorded in notifications_sent within timeout")
}
