package internalhttp

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// loggingMiddleware logs every served request in the format required by the
// HW12 spec, e.g.:
//
//	66.249.65.3 [25/Feb/2020:19:11:24 +0600] GET /hello?q=1 HTTP/1.1 200 30 "Mozilla/5.0"
func loggingMiddleware(logger Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		line := fmt.Sprintf(
			"%s [%s] %s %s %s %d %d %q",
			clientIP(r),
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			requestURI(r),
			r.Proto,
			rw.status,
			latency.Milliseconds(),
			r.UserAgent(),
		)
		logger.Info(line)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func requestURI(r *http.Request) string {
	if r.URL.RawQuery != "" {
		return r.URL.Path + "?" + r.URL.RawQuery
	}
	return r.URL.Path
}
