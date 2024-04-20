package models

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/natac13/go-chirpy/internal/auth"
	"github.com/natac13/go-chirpy/internal/database"
	"github.com/natac13/go-chirpy/internal/response"
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

func HandleGetChirps(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chirps, err := db.GetChirps()
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.RespondWithJSON(w, http.StatusOK, chirps)
	}
}

func HandleGetChirp(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		chirps, err := db.GetChirps()
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for _, chirp := range chirps {
			if chirp.Id == id {
				response.RespondWithJSON(w, http.StatusOK, chirp)
				return
			}
		}

		response.RespondWithError(w, http.StatusNotFound, "Chirp not found")
	}

}

func HandleCreateChirp(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userId, err := auth.ValidateToken(r)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		decoder := json.NewDecoder(r.Body)
		var chirpRequest ChirpRequest
		err = decoder.Decode(&chirpRequest)
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if len(chirpRequest.Body) > 140 {
			response.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
			return
		}

		chirp, err := db.CreateChirp(chirpRequest.Body, userId)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.RespondWithJSON(w, http.StatusCreated, chirp)
	}
}
