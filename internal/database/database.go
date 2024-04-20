package database

import (
	"encoding/json"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  sync.RWMutex
}

type Chirp struct {
	AuthorId int    `json:"author_id"`
	Body     string `json:"body"`
	Id       int    `json:"id"`
}

type User struct {
	Email    string `json:"email"`
	Id       int    `json:"id"`
	Password string `json:"password"`
}

type RevokedToken struct {
	RevokedAt time.Time `json:"revoked_at"`
}

type DBStructure struct {
	Chirps        map[int]Chirp           `json:"chirps"`
	Users         map[int]User            `json:"users"`
	RevokedTokens map[string]RevokedToken `json:"revoked_tokens"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  sync.RWMutex{},
	}

	if err := db.ensureDB(); err != nil {
		return nil, err
	}

	return db, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	// create the database file if it doesn't exist
	_, err := os.ReadFile(db.path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(db.path, []byte("{}"), 0666); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	data := DBStructure{
		Chirps:        map[int]Chirp{},
		Users:         map[int]User{},
		RevokedTokens: map[string]RevokedToken{},
	}

	file, err := os.ReadFile(db.path)
	if err != nil {
		return data, err
	}

	if err := json.Unmarshal(file, &data); err != nil {
		return data, err
	}

	return data, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	if err := os.WriteFile(db.path, data, 0666); err != nil {
		return err
	}

	return nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, userId int) (Chirp, error) {

	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp := Chirp{
		Id:       len(data.Chirps) + 1,
		Body:     body,
		AuthorId: userId,
	}

	data.Chirps[chirp.Id] = chirp

	if err := db.writeDB(data); err != nil {
		return chirp, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {

	data, err := db.loadDB()
	if err != nil {
		slog.Error("DATABASE - Error getting chirps", "error", err)
		return nil, err
	}

	if len(data.Chirps) == 0 {
		return []Chirp{}, nil
	}

	chirps := []Chirp{}
	for _, chirp := range data.Chirps {
		chirps = append(chirps, chirp)
	}

	sort.Slice(chirps, func(i, y int) bool {
		return chirps[i].Id < chirps[y].Id
	})

	return chirps, nil
}

func (db *DB) GetChirpById(chirpId int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := data.Chirps[chirpId]
	if !ok {
		return Chirp{}, nil
	}

	return chirp, nil
}

func (db *DB) DeleteChirp(chirpId int) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}

	delete(data.Chirps, chirpId)

	if err := db.writeDB(data); err != nil {
		return err
	}

	return nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if db.checkDuplicateEmail(email) {
		return User{}, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{
		Id:       len(data.Users) + 1,
		Email:    email,
		Password: string(hash),
	}

	data.Users[user.Id] = user

	if err := db.writeDB(data); err != nil {
		return user, err
	}

	return user, nil
}

func (db *DB) UpdateUser(userId int, email, password string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, err := db.GetUserById(userId)
	if err != nil {
		return User{}, err
	}

	if email != "" {
		user.Email = email
	}

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}

		user.Password = string(hash)
	}

	data.Users[user.Id] = user

	if err := db.writeDB(data); err != nil {
		return user, err
	}

	return user, nil
}

func (db *DB) RevokeToken(token string) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}

	data.RevokedTokens[token] = RevokedToken{
		RevokedAt: time.Now().UTC(),
	}

	if err := db.writeDB(data); err != nil {
		return err
	}

	return nil
}

func (db *DB) IsTokenRevoked(token string) bool {
	data, err := db.loadDB()
	if err != nil {
		return true
	}

	_, ok := data.RevokedTokens[token]
	return ok
}

func (db *DB) checkDuplicateEmail(email string) bool {
	data, err := db.loadDB()
	if err != nil {
		return true
	}

	for _, user := range data.Users {
		if user.Email == email {
			return true
		}
	}

	return false
}

func (db *DB) VerifyPassword(email, password string) (User, error) {
	user, err := db.getUserByEmail(email)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) getUserByEmail(email string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range data.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, nil
}

func (db *DB) GetUserById(id int) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range data.Users {
		if user.Id == id {
			return user, nil
		}
	}

	return User{}, nil
}
