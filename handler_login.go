package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	params := &parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
	}

	dbUsers, err := cfg.DB.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve users")
		return
	}

	for _, dbUser := range dbUsers {

		claims := jwt.RegisteredClaims{}
		if params.ExpiresInSeconds == 0 {
			claims = jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				Issuer:    "chirpy",
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				Subject:   strconv.Itoa(dbUser.ID),
			}
		} else {
			claims = jwt.RegisteredClaims{
				Issuer:    "chirpy",
				IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(params.ExpiresInSeconds) * time.Second)),
				Subject:   strconv.Itoa(dbUser.ID),
			}
		}

		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := newToken.SignedString([]byte(cfg.jwtSecret))
		if err != nil {
			log.Println("Error Creating jwt")
		}

		dat := make([]byte, 32)
		_, err = rand.Read(dat)
		if err != nil {
			log.Println("error creating refresh token")
		}

		refreshToken := hex.EncodeToString(dat)
		err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(params.Password))
		if dbUser.Email != params.Email || err != nil {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		} else {

			err = cfg.DB.SaveRefreshToken(dbUser.ID, refreshToken)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't save refreshToken")
				return
			}
			respondWithJSON(w, http.StatusOK, User{
				ID:           dbUser.ID,
				Email:        dbUser.Email,
				Token:        tokenString,
				RefreshToken: refreshToken,
			})
		}
	}

}
