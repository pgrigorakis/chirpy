package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/pgrigorakis/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT: %w", err)
		return
	}
	userID, err := auth.ValidateJWT(accessToken, cfg.jwtToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't retrieve chirps: %w", err)
		return
	}
	if chirp.UserID == userID {
		respondWithError(w, http.StatusForbidden, "This user is not allowed to delete this chirp.", err)
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirps: %w", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
