package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Users         map[int]User            `json:"users"`
	Chirps        map[int]Chirp           `json:"chirps"`
	RefreshTokens map[string]RefreshToken `json:"refresh_tokens"`
}

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) GetUsers() (map[int]User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	return dbStructure.Users, nil
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	log.Printf("Creating user with email: %s", email)
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return User{}, errors.New("email already in use")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	id := len(dbStructure.Users) + 1
	user := User{
		ID:       id,
		Email:    email,
		Password: string(hashedPassword),
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpdateUser(id int, email, password string) (User, error) {
	log.Printf("Updating email and password of user with id: %d", id)
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return User{}, errors.New("email already in use")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	user := User{
		ID:       id,
		Email:    email,
		Password: string(hashedPassword),
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) Revoke(refreshToken string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	delete(dbStructure.RefreshTokens, refreshToken)
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SaveRefreshToken(userID int, token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	if dbStructure.RefreshTokens == nil {
		dbStructure.RefreshTokens = make(map[string]RefreshToken)
	}
	refreshToken := RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	dbStructure.RefreshTokens[token] = refreshToken

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UserForRefreshToken(token string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	refreshToken, ok := dbStructure.RefreshTokens[token]
	if !ok {
		return User{}, errors.New("resource doesnt exist")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return User{}, errors.New("token expired")
	}

	user, err := db.GetUser(refreshToken.UserID)
	if err != nil {
		return User{}, err
	}

	return user, nil

}

func (db *DB) GetUser(userID int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[userID]
	if !ok {
		return User{}, errors.New("user doesn't exist")
	}

	return user, nil
}
