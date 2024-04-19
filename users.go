package main

import (
	"encoding/json"
	"net/http"

	"github.com/natac13/go-chirpy/internal/database"
)

type UserRequest struct {
	Email string `json:"email"`
}

func handleCreateUser(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var userRequest UserRequest
		err := decoder.Decode(&userRequest)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		user, err := db.CreateUser(userRequest.Email)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusCreated, user)
	}
}
