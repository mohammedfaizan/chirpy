package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {

	type response struct {
		Token string `json:"token"`
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	user, err := cfg.DB.UserForRefreshToken(token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get refresh token user")
		return
	}

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		Subject:   strconv.Itoa(user.ID),
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := newToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		log.Println("Error Creating jwt")
		return
	}
	cfg.DB.SaveRefreshToken(user.ID, tokenString)
	respondWithJSON(w, http.StatusOK, response{
		Token: tokenString,
	})
}
