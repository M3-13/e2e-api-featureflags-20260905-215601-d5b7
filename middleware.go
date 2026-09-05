package main

import "net/http"

func withLogging(next http.Handler) http.Handler {
	return next
}
