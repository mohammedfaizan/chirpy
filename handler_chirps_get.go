package main

import (
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {

	authid := r.URL.Query().Get("author_id")
	sortType := r.URL.Query().Get("sort")
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {

		switch {
		case authid != "":
			id, err := strconv.Atoi(authid)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "invalid author id")
				return
			}
			if dbChirp.AuthorId == id {
				chirps = append(chirps, Chirp(dbChirp))
			}

		default:
			chirps = append(chirps, Chirp(dbChirp))
		}
	}

	switch {
	case sortType == "asc":
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	case sortType == "desc":
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	default:
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}
