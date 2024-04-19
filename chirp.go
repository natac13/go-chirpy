package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/natac13/go-chirpy/internal/database"
)

type ChirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

type ChirpRequest struct {
	Body string `json:"body"`
}

func cleanChirpMessage(m string) string {
	// bad words:
	// kerfuffle
	// sharbert
	// fornax
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
		chirps, err := db.GetChirps()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, chirps)
	}
}

func handleGetChirp(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		chirps, err := db.GetChirps()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		slog.Info("Getting chirp", "id", id, "chirps", chirps)

		for _, chirp := range chirps {
			if chirp.Id == id {
				respondWithJSON(w, http.StatusOK, chirp)
				return
			}
		}

		respondWithError(w, http.StatusNotFound, "Chirp not found")
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
