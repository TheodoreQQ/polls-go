package repository

import (
	"database/sql"

	"github.com/lib/pq"
)

// struct handles db operations for user authentication and management
type AuthRepository struct {
	DB *sql.DB
}

// inserts new user record in the db
func (r *AuthRepository) CreateUser(username, hashedPassword string) error {
	var userID int

	queryUser := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`

	err := r.DB.QueryRow(queryUser, username, hashedPassword).Scan(&userID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return ErrUserAlreadyExists
			}
		}
		return err
	}
	return nil
}

// retrieves a user from database based on their unique username
func (r *AuthRepository) GetUser(username string) (int, string, error) {
	var storedID int
	var storedHash string

	query := `SELECT id, password_hash FROM users WHERE username = $1`

	err := r.DB.QueryRow(query, username).Scan(&storedID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", ErrNotFound
		}
		return 0, "", err
	}

	return storedID, storedHash, nil
}
