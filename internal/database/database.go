package database

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  sync.RWMutex
}

type Chirp struct {
	Body string `json:"body"`
	Id   int    `json:"id"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	slog.Info("New database", "path", path)
	db := &DB{
		path: path,
		mux:  sync.RWMutex{},
	}

	slog.Info("Ensuring database exists", "path", path)
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
			slog.Info("Creating new database file", "path", db.path)
			if err := os.WriteFile(db.path, []byte("{}"), 0666); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	slog.Info("Database ready file at", "path", db.path)
	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	data := DBStructure{
		Chirps: map[int]Chirp{},
	}

	slog.Info("Loading database file", "path", db.path)
	file, err := os.ReadFile(db.path)
	if err != nil {
		return data, err
	}

	if err := json.Unmarshal(file, &data); err != nil {
		return data, err
	}

	slog.Info("Database loaded into memory", "path", db.path, "data", data)

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
func (db *DB) CreateChirp(body string) (Chirp, error) {

	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp := Chirp{
		Id:   len(data.Chirps) + 1,
		Body: body,
	}

	data.Chirps[chirp.Id] = chirp

	if err := db.writeDB(data); err != nil {
		return chirp, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {

	slog.Info("DATABASE - Getting chirps")

	data, err := db.loadDB()
	if err != nil {
		slog.Error("DATABASE - Error getting chirps", "error", err)
		return nil, err
	}

	slog.Info("DATABASE - Got chirps", "chirps", data.Chirps)

	if len(data.Chirps) == 0 {
		return []Chirp{}, nil
	}

	chirps := []Chirp{}
	for _, chirp := range data.Chirps {
		chirps = append(chirps, chirp)
	}

	// sort.Slice(chirps, func(i, y int) bool {
	// 	return chirps[i].Id < chirps[y].Id
	// })

	slog.Info("DATABASE - Returning chirps", "chirps", chirps)
	return chirps, nil
}
