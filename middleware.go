package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

// accessLogger writes one line per request to stdout. It is a package-level
// variable so tests can swap its output.
var accessLogger = log.New(os.Stdout, "", log.LstdFlags)

// statusRecorder wraps http.ResponseWriter and remembers the status code that
// was actually written, defaulting to 200 when the handler never calls
// WriteHeader explicitly.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		duration := time.Since(start)

		// Log only the method, r.URL.Path (never the query string) and status.
		accessLogger.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, duration)
	})
}
