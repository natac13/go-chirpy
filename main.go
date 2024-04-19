package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func main() {
	router := http.NewServeMux()
	config := &apiConfig{
		fileserverHits: 0,
	}

	staticFiles := http.FileServer(http.Dir("."))

	router.Handle("/app/*", http.StripPrefix("/app", config.metricsHitMiddleware(staticFiles)))
	router.HandleFunc("GET /api/healthz", handleHealthz)
	router.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		output := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has beed visited %d times!</p></body></html>", config.fileserverHits)
		w.Write([]byte(output))
	})
	router.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		config.fileserverHits = 0

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Hits reset\n"))
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: middlewareCors(router),
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server: ", "error", err)
	}
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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
