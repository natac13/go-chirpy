package models

import (
	"encoding/json"
	"net/http"

	"github.com/natac13/go-chirpy/internal/auth"
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

type UserResponse struct {
	Email        string `json:"email"`
	Id           int    `json:"id"`
	Password     string `json:"-"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
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

		response.RespondWithJSON(w, http.StatusCreated, UserResponse{
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

		accessToken, err := auth.GetAccessToken(user.Id)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "Error generating access token")
			return
		}

		refreshToken, err := auth.GetRefreshToken(user.Id)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "Error generating refresh token")
			return
		}

		response.RespondWithJSON(w, http.StatusOK, UserResponse{
			Email:        user.Email,
			Id:           user.Id,
			Token:        accessToken,
			RefreshToken: refreshToken,
		})
	}
}

type UserUpdateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func HandleUpdateUser(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := auth.ValidateToken(r)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		decoder := json.NewDecoder(r.Body)
		var userUpdateRequest UserUpdateRequest
		err = decoder.Decode(&userUpdateRequest)
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		user, err := db.UpdateUser(userId, userUpdateRequest.Email, userUpdateRequest.Password)

		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.RespondWithJSON(w, http.StatusOK, UserResponse{
			Email: user.Email,
			Id:    user.Id,
		})

	}
}
