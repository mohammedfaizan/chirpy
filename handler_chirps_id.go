package main

import (
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerGetChirpsByID(w http.ResponseWriter, r *http.Request) {

	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp ID")
		return
	}

	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "invalid chirp id")
	}
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:   dbChirp.ID,
			Body: dbChirp.Body,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	if id > len(chirps) {
		respondWithError(w, http.StatusNotFound, "chirp not present")
		return
	}

	chirp := Chirp{
		ID:   chirps[id-1].ID,
		Body: chirps[id-1].Body,
	}

	respondWithJSON(w, http.StatusOK, chirp)
}
