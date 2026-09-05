package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// captureAccessLog redirects the access logger into a buffer and restores the
// previous output when the test finishes.
func captureAccessLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := accessLogger.Writer()
	accessLogger.SetOutput(buf)
	t.Cleanup(func() { accessLogger.SetOutput(prev) })
	return buf
}

var durationPattern = regexp.MustCompile(`\d+(\.\d+)?(ns|µs|ms|s)`)

func TestWithLoggingLogsMethodPathStatusDuration(t *testing.T) {
	buf := captureAccessLog(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	h := withLogging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/some-key/evaluate?user=42", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 passthrough, got %d", rr.Code)
	}

	logLine := buf.String()
	if !strings.Contains(logLine, "GET") {
		t.Fatalf("expected method in log line, got %q", logLine)
	}
	if !strings.Contains(logLine, "/flags/some-key/evaluate") {
		t.Fatalf("expected path in log line, got %q", logLine)
	}
	if !strings.Contains(logLine, "201") {
		t.Fatalf("expected status in log line, got %q", logLine)
	}
	if !durationPattern.MatchString(logLine) {
		t.Fatalf("expected a duration in log line, got %q", logLine)
	}
}

func TestWithLoggingNeverLogsQueryStringOrUserID(t *testing.T) {
	buf := captureAccessLog(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := withLogging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/mykey/evaluate?user=secretuser123", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 passthrough, got %d", rr.Code)
	}

	logLine := buf.String()
	if strings.Contains(logLine, "?") {
		t.Fatalf("log line must not contain a query string, got %q", logLine)
	}
	if strings.Contains(logLine, "secretuser123") {
		t.Fatalf("log line must not contain the user id, got %q", logLine)
	}
	if strings.Contains(logLine, "user=") {
		t.Fatalf("log line must not contain the query parameter, got %q", logLine)
	}
}

func TestWithLoggingDefaultsStatusTo200(t *testing.T) {
	buf := captureAccessLog(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	h := withLogging(next)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	logLine := buf.String()
	if !strings.Contains(logLine, "200") {
		t.Fatalf("expected default 200 in log line, got %q", logLine)
	}
}
