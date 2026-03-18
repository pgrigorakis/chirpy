package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func cleanText(input string) (string, error) {

	bannedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	words := strings.Split(input, " ")

	for i, word := range words {
		if _, ok := bannedWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " "), nil

}

func handlerValidateChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", err)
		return
	}

	cleanedText, err := cleanText(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't clean text", err)
		return
	}

	respondWithJSON(w, 200, responseBody{
		Valid:       true,
		CleanedBody: cleanedText,
	})
}
