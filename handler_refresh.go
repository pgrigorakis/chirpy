package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/pgrigorakis/chirpy/internal/auth"
	"github.com/pgrigorakis/chirpy/internal/database"
)

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	type responseBody struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token: %w", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found", err)
		return
	}

	accessToken, err := auth.MakeJWT(user, cfg.jwtToken, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token: %w", err)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerTokenRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get token: %w", err)
		return
	}
	err = cfg.db.RevokeToken(r.Context(), database.RevokeTokenParams{
		UpdatedAt: time.Now(),
		RevokedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Token: refreshToken,
	})

	w.WriteHeader(http.StatusNoContent)

}
