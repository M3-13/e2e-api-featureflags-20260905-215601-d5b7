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

// sanitizePath replaces every control character (byte < 32) and DEL (127)
// with '?', so raw control bytes never reach the access log.
func sanitizePath(path string) string {
	b := []byte(path)
	for i, c := range b {
		if c < 32 || c == 127 {
			b[i] = '?'
		}
	}
	return string(b)
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			if p := recover(); p != nil {
				rec.status = http.StatusInternalServerError
				writeError(rec, http.StatusInternalServerError, "internal server error")
				// Log the panic value internally only — never in the response body.
				accessLogger.Printf("panic: %v", p)
			}

			// The normal access log line is always written, including on panic
			// (with the status forced to 500 above).
			duration := time.Since(start)
			accessLogger.Printf("%s %s %d %s", r.Method, sanitizePath(r.URL.Path), rec.status, duration)
		}()

		next.ServeHTTP(rec, r)
	})
}
