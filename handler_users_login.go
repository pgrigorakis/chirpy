package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pgrigorakis/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type requestBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type responseBody struct {
		User
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters: %w", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	passwordMatches, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !passwordMatches {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	if params.ExpiresInSeconds <= 0 || params.ExpiresInSeconds < 3600 {
		params.ExpiresInSeconds = 3600
	}
	expiresInSecondsTime := time.Second * time.Duration(params.ExpiresInSeconds)

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtToken, expiresInSecondsTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token: %w", err)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email},
		Token: accessToken,
	})
}
