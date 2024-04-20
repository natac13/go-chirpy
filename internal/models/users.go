package models

import (
	"encoding/json"
	"net/http"

	"github.com/natac13/go-chirpy/internal/database"
	"github.com/natac13/go-chirpy/internal/response"
)

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func HandleCreateUser(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var userRequest UserRequest
		err := decoder.Decode(&userRequest)
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		user, err := db.CreateUser(userRequest.Email, userRequest.Password)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.RespondWithJSON(w, http.StatusCreated, struct {
			Email string `json:"email"`
			Id    int    `json:"id"`
		}{
			Email: user.Email,
			Id:    user.Id,
		})
	}
}

func HandleUserLogin(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var loginRequest LoginRequest
		err := decoder.Decode(&loginRequest)
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		user, err := db.VerifyPassword(loginRequest.Email, loginRequest.Password)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		response.RespondWithJSON(w, http.StatusOK, struct {
			Email string `json:"email"`
			Id    int    `json:"id"`
		}{
			Email: user.Email,
			Id:    user.Id,
		})
	}
}
