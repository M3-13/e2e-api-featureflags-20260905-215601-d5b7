package main

import (
	"bytes"
	"encoding/json"
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

func TestWithLoggingRecoversPanic(t *testing.T) {
	buf := captureAccessLog(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := withLogging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/mykey", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on panic, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json body: %v", err)
	}
	if body["error"] != "internal server error" {
		t.Fatalf("expected generic error message, got %q", body["error"])
	}
	if len(body) != 1 {
		t.Fatalf("error object must contain exactly the error field, got %v", body)
	}
	if strings.Contains(rr.Body.String(), "boom") {
		t.Fatalf("panic value must not leak into the response body, got %q", rr.Body.String())
	}

	logLine := buf.String()
	if !strings.Contains(logLine, "panic:") {
		t.Fatalf("expected 'panic:' in the log, got %q", logLine)
	}
	if !strings.Contains(logLine, "500") {
		t.Fatalf("expected status 500 in the log line, got %q", logLine)
	}
	if !strings.Contains(logLine, "GET") {
		t.Fatalf("expected method in the log line, got %q", logLine)
	}
	if !strings.Contains(logLine, "/flags/mykey") {
		t.Fatalf("expected path in the log line, got %q", logLine)
	}
	if !durationPattern.MatchString(logLine) {
		t.Fatalf("expected a duration in the log line, got %q", logLine)
	}
}

func TestWithLoggingSanitizesPath(t *testing.T) {
	buf := captureAccessLog(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := withLogging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/x%0atest", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	logLine := buf.String()
	if !strings.Contains(logLine, "/flags/x?test") {
		t.Fatalf("expected sanitized path in the log line, got %q", logLine)
	}
	if strings.Contains(logLine, "/flags/x\ntest") {
		t.Fatalf("raw control character leaked into the log line %q", logLine)
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
