package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type EventRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {

	var req EventRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad request")
		return
	}

	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user, err := cfg.DB.UpgradeUserRed(req.Data.UserID)
	if err != nil {

		if errors.Is(err, fmt.Errorf("user with id doesn't exists")) {
			respondWithError(w, http.StatusNotFound, "couldn'nt find user")
			return
		}
	}
	log.Printf("handler webhook")
	log.Printf("%v", user.IsChirpyRed)
	w.WriteHeader(http.StatusNoContent)
}
