package main

import (
	"net/http"
	"strings"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	err := cfg.DB.Revoke(token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
