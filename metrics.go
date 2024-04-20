package main

import (
	"fmt"
	"net/http"
)

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	// set the content type
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// write the response
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits int
}

func (a *apiConfig) metricsHitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits = a.fileserverHits + 1
		next.ServeHTTP(w, r)
	})
}

func handleAdminMetric(c *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		content := "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>"

		output := fmt.Sprintf(content, c.fileserverHits)
		w.Write([]byte(output))
	}
}

func handleReset(c *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits = 0

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Hits reset\n"))
	}
}
