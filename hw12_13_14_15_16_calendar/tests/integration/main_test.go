package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
)

const (
	defaultCalendarURL = "http://localhost:8888"
	defaultDBDSN       = "postgres://postgres:password@localhost:5435/calendar?sslmode=disable"
	waitTimeout        = 60 * time.Second
	waitInterval       = 2 * time.Second
	httpClientTimeout  = 10 * time.Second
)

var (
	calendarURL string
	testDB      *sql.DB
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	calendarURL = envOrDefault("CALENDAR_HTTP_URL", defaultCalendarURL)
	dbDSN := envOrDefault("INTEGRATION_DB_DSN", defaultDBDSN)

	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		fmt.Printf("failed to open db: %v\n", err)
		return 1
	}
	testDB = db
	defer func() { _ = db.Close() }()

	if err := waitForHTTP(calendarURL + "/events/day?date=2024-01-01"); err != nil {
		fmt.Printf("calendar not ready: %v\n", err)
		return 1
	}

	return m.Run()
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func waitForHTTP(url string) error {
	client := &http.Client{Timeout: httpClientTimeout}
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(waitInterval)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}

// apiRequest performs an HTTP request to the calendar API.
func apiRequest(method, path, body string) (*http.Response, error) {
	client := &http.Client{Timeout: httpClientTimeout}
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, calendarURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return client.Do(req)
}
