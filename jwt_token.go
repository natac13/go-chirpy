package main

import (
	"net/http"

	"github.com/natac13/go-chirpy/internal/auth"
	"github.com/natac13/go-chirpy/internal/database"
	"github.com/natac13/go-chirpy/internal/response"
)

func RevokeTokenHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenWithBearer := r.Header.Get("Authorization")
		if tokenWithBearer == "" {
			response.RespondWithError(w, http.StatusUnauthorized, "No token provided")
			return
		}

		token := tokenWithBearer[len("Bearer "):]
		if token == "" {
			response.RespondWithError(w, http.StatusUnauthorized, "No token provided")
			return
		}

		if err := db.RevokeToken(token); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Token revoked"})
	}
}

func RefreshTokenHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.RespondWithError(w, http.StatusUnauthorized, "No token provided")
			return
		}

		userId, tokenString, err := auth.ValidateRefreshToken(authHeader)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		if db.IsTokenRevoked(tokenString) {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		if err := db.RevokeToken(tokenString); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		user, err := db.GetUserById(userId)
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		newToken, err := auth.GetAccessToken(user.Id)

		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.RespondWithJSON(w, http.StatusOK, map[string]string{"token": newToken})
	}
}
