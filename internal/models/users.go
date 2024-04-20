package models

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/natac13/go-chirpy/internal/database"
	"github.com/natac13/go-chirpy/internal/response"
)

type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
}

type UserResponse struct {
	Email    string `json:"email"`
	Id       int    `json:"id"`
	Password string `json:"-"`
	Token    string `json:"token,omitempty"`
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

		// get the expiry in seconds which defaults to 24 hours
		// and has a max of 24 hours
		var expiry time.Duration
		if loginRequest.ExpiresInSeconds > 0 {
			expiry = time.Duration(loginRequest.ExpiresInSeconds) * time.Second
			if expiry > 24*time.Hour {
				expiry = 24 * time.Hour
			}
		} else {
			expiry = 24 * time.Hour // default expiry of 24 hours
		}

		claims := jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Subject:   strconv.Itoa(user.Id), // convert int to string
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiry)),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		secret := os.Getenv("JWT_SECRET")
		ss, err := token.SignedString([]byte(secret))
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "Error signing token")
			return
		}

		response.RespondWithJSON(w, http.StatusOK, UserResponse{
			Email: user.Email,
			Id:    user.Id,
			Token: ss,
		})
	}
}

type UserUpdateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func HandleUpdateUser(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.RespondWithError(w, http.StatusUnauthorized, "No authorization header")
			return
		}
		slog.Info("\nAuthHeader: ", "header", authHeader)

		tokenString := authHeader[len("Bearer "):]
		slog.Info("\nTokenString", "token", tokenString)
		claims := jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		slog.Info("\nClaims: ", "claims", claims)

		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		slog.Info("\nToken: ", "token", token, "valid", token.Valid)

		if !token.Valid {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		userIdString := claims.Subject
		slog.Info("\nUserIdString: ", "userIdString", userIdString)
		// convert string to int
		userId, err := strconv.Atoi(userIdString)
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "Error converting user id")
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
