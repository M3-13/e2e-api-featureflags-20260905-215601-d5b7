package main

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

// withAuth guards the /flags routes behind a bearer token. When API_TOKEN is
// empty (dev/localhost mode) every request passes through unauthenticated.
// When it is set, requests whose path is /flags or below /flags must carry
// "Authorization: Bearer <API_TOKEN>"; anything else (including /healthz)
// remains open.
func withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := os.Getenv("API_TOKEN")
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.HasPrefix(r.URL.Path, "/flags") {
			next.ServeHTTP(w, r)
			return
		}

		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		provided := strings.TrimPrefix(header, prefix)
		if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
