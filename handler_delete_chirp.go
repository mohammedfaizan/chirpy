package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	authHead := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHead, "Bearer ")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			respondWithError(w, http.StatusUnauthorized, "Invalid Token String")
			return
		}
		respondWithError(w, http.StatusForbidden, "Invalid Token")
		return
	}

	if !token.Valid {
		respondWithError(w, http.StatusForbidden, "invalid token")
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error recieving subject")
		return
	}

	log.Printf("Token is valid. Issuer: %v\n", subject)

	id, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid User")
		return
	}

	chirpId := r.PathValue("chirpID")
	if chirpId == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp ID")
		return
	}

	chirpid, err := strconv.Atoi(chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "invalid chirp id")
		return
	}

	dbChirp, err := cfg.DB.GetChirp(chirpid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "chirp not present")
		return
	}

	if id == dbChirp.AuthorId {
		err = cfg.DB.DeleteChirp(chirpid)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp")
			return
		}

		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	respondWithError(w, http.StatusForbidden, "Forbidden")

}
