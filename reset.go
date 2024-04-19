package main

import "net/http"

func handleReset(c *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits = 0

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Hits reset\n"))
	}
}
