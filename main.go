package main

import (
	"log/slog"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	router.Handle("/app/*", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	router.HandleFunc("/healthz", handleHealthz)

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
