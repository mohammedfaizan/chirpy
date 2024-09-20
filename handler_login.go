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

	log.Printf("handler login")
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	dbUser, err := cfg.DB.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve users")
		return
	}

	claims := jwt.RegisteredClaims{}
	claims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Subject:   strconv.Itoa(dbUser.ID),
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
		respondWithJSON(w, http.StatusOK, response{
			User: User{
				ID:          dbUser.ID,
				Email:       dbUser.Email,
				IsChirpyRed: dbUser.IsChirpyRed,
			},
			Token:        tokenString,
			RefreshToken: refreshToken,
		})
	}

}
