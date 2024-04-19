package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/natac13/go-chirpy/internal/database"
)

type ChirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

type ChirpRequest struct {
	Body string `json:"body"`
}

// bad words:
// kerfuffle
// sharbert
// fornax

func cleanChirpMessage(m string) string {
	listOfBadWords := []string{"kerfuffle", "sharbert", "fornax"}
	listOfWords := strings.Split(m, " ")
	replacement := "****"

	newWords := []string{}

	for _, word := range listOfWords {
		if !slices.Contains(listOfBadWords, strings.ToLower((word))) {
			newWords = append(newWords, word)
		} else {
			newWords = append(newWords, replacement)
		}
	}

	return strings.Join(newWords, " ")

}

func handleGetChirps(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Getting chirps")
		chirps, err := db.GetChirps()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		slog.Info("Got chirps", "chirps", chirps)

		respondWithJSON(w, http.StatusOK, chirps)
	}
}

func handleCreateChirp(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var chirpRequest ChirpRequest
		err := decoder.Decode(&chirpRequest)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if len(chirpRequest.Body) > 140 {
			respondWithError(w, http.StatusBadRequest, "Chirp is too long")
			return
		}

		chirp, err := db.CreateChirp(chirpRequest.Body)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusCreated, chirp)
	}
}
