package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/natac13/go-chirpy/internal/database"
)

const (
	databasePath = "database.json"
)

func main() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if &dbg != nil && *dbg {
		slog.Info("Debug mode enabled. Deleting database file.", "path", databasePath)
		os.Remove(databasePath)
	}

	router := http.NewServeMux()
	config := &apiConfig{
		fileserverHits: 0,
	}

	staticFiles := http.FileServer(http.Dir("."))

	db, err := database.NewDB(databasePath)
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
	router.HandleFunc("GET /api/chirps/{id}", handleGetChirp(db))

	router.HandleFunc("POST /api/users", handleCreateUser(db))

	server := http.Server{
		Addr:    ":8080",
		Handler: middlewareCors(router),
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server: ", "error", err)
	}
}
