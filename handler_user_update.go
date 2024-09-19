package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
}

func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := &parameters{}
	decoder.Decode(&params)
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
		respondWithError(w, http.StatusUnauthorized, "Invalid Token")
		return
	}

	if !token.Valid {
		respondWithError(w, http.StatusBadRequest, "invalid token")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error recieving subject")
	}

	log.Printf("Token is valid. Issuer: %v\n", subject)

	id, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid User")
	}

	_, err = cfg.DB.UpdateUser(id, params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:    id,
		Email: params.Email,
	})
}
