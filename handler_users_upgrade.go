package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/pgrigorakis/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUsersUpgrade(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type requestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	headerApiKey, err := auth.GetPolkaKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't read header: %w", err)
		return
	}

	if headerApiKey != os.Getenv("POLKA_KEY") {
		respondWithError(w, http.StatusUnauthorized, "Incorrect API key", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters: %w", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	_, err = cfg.db.UpdateUserToRed(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
			return
		}
		respondWithError(w, http.StatusNotFound, "Couldn't update user: %w", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
