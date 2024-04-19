package main

import (
	"slices"
	"strings"
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
