package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/natac13/go-chirpy/internal/database"
	"github.com/natac13/go-chirpy/internal/response"
)

type PolkaRequest struct {
	Event string         `json:"event"`
	Data  PolkaEventData `json:"data"`
}

type PolkaEventData struct {
	UserID int `json:"user_id"`
}

func handlePolkaWebhook(db *database.DB) http.HandlerFunc {
	apiKey := os.Getenv("POLKA_API_KEY")
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.RespondWithError(w, http.StatusUnauthorized, "No token provided")
			return
		}

		authHeaderKey := strings.TrimLeft(authHeader, "ApiKey ")
		if authHeaderKey != apiKey {
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		decoder := json.NewDecoder(r.Body)
		var polkaRequest PolkaRequest
		err := decoder.Decode(&polkaRequest)
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		event := polkaRequest.Event
		switch event {
		case "user.upgraded":
			err := db.UpgradeToChirpyRed(polkaRequest.Data.UserID)
			if err != nil {
				if strings.ToLower(err.Error()) == "user not found" {
					response.RespondWithError(w, http.StatusNotFound, "User not found")
					break
				}
			}

			response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User upgraded"})
			break
		default:
			response.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Event not supported"})
		}
	}
}
