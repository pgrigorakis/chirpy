package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgrigorakis/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type responseBody struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleanedText, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't clean text", err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedText,
		UserID: params.UserID},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, responseBody{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID},
	})
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetAllChirpsByCreateDate(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't retrieve chirps: %w", err)
		return
	}

	var chirps []Chirp
	// chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirp := Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}
		chirps = append(chirps, chirp)
	}

	respondWithJSON(w, http.StatusOK, dbChirps)
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	cleaned := cleanText(body)
	return cleaned, nil
}

func cleanText(input string) string {
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

	return strings.Join(words, " ")
}
