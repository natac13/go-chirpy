package main

import (
	"log/slog"
	"net/http"

	"github.com/natac13/go-chirpy/internal/database"
)

func main() {
	router := http.NewServeMux()
	config := &apiConfig{
		fileserverHits: 0,
	}

	staticFiles := http.FileServer(http.Dir("."))

	db, err := database.NewDB("database.json")
	if err != nil {
		slog.Error("Error opening database: ", "error", err)
		panic("Error opening database")
	}

	router.Handle("/app/*", http.StripPrefix("/app", config.metricsHitMiddleware(staticFiles)))
	router.HandleFunc("GET /api/healthz", handleHealthz)
	router.HandleFunc("GET /admin/metrics", handleAdminMetric(config))
	router.HandleFunc("/api/reset", handleReset(config))

	router.HandleFunc("POST /api/chirps", handleCreateChirp(db))
	router.HandleFunc("GET /api/chirps", handleGetChirps(db))

	server := http.Server{
		Addr:    ":8080",
		Handler: middlewareCors(router),
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server: ", "error", err)
	}
}
