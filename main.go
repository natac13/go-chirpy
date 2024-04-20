package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/natac13/go-chirpy/internal/database"
	"github.com/natac13/go-chirpy/internal/models"
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
	godotenv.Load()

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

	router.HandleFunc("POST /api/chirps", models.HandleCreateChirp(db))
	router.HandleFunc("GET /api/chirps", models.HandleGetChirps(db))
	router.HandleFunc("GET /api/chirps/{id}", models.HandleGetChirp(db))

	router.HandleFunc("POST /api/users", models.HandleCreateUser(db))
	router.HandleFunc("POST /api/login", models.HandleUserLogin(db))
	router.HandleFunc("PUT /api/users", models.HandleUpdateUser(db))

	router.HandleFunc("POST /api/revoke", RevokeTokenHandler(db))
	router.HandleFunc("POST /api/refresh", RefreshTokenHandler(db))

	server := http.Server{
		Addr:    ":8080",
		Handler: middlewareCors(router),
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Error starting server: ", "error", err)
	}
}
