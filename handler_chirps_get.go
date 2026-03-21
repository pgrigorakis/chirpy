package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/pgrigorakis/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	authorID, err := uuid.Parse(r.URL.Query().Get("author_id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse author id: %w", err)
	}

	var dbChirps []database.Chirp
	if authorID == uuid.Nil {
		dbChirps, err = cfg.db.GetAllChirpsByCreateDate(r.Context())
	} else {
		dbChirps, err = cfg.db.GetChirpsByAuthor(r.Context(), authorID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps: %w", err)
		return
	}

	var chirps []Chirp // or chirps := []Chirp{}
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

func (cfg *apiConfig) handlerChirpsGetByID(w http.ResponseWriter, r *http.Request) {
	type responseBody struct {
		Chirp
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't retrieve chirps: %w", err)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID},
	})
}
