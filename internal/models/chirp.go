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
		s := r.URL.Query().Get("author_id")
		sorting := r.URL.Query().Get("sort")

		authorId, err := strconv.Atoi(s)
		if err != nil {
			authorId = 0
		}

		if sorting == "" {
			sorting = "asc"
		}

		chirps, err := db.GetChirps(authorId, sorting)
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
		chirps, err := db.GetChirps(0, "asc")
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

func HandleDeleteChirp(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := auth.ValidateToken(r)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, "Invalid chirp id")
			return
		}

		chirp, err := db.GetChirpById(id)
		if err != nil {
			response.RespondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}

		if chirp.AuthorId != userId {
			response.RespondWithError(w, http.StatusForbidden, "You are not the author of this chirp")
			return
		}

		if err := db.DeleteChirp(id); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Chirp deleted"})
	}
}
