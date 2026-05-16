package handlers

import (
	"database/sql"
	"log"
	"os"
)

func ConnectDB() *sql.DB {

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		log.Fatal("URL is empty")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	InitDB(db)
	return db
}

func InitDB(db *sql.DB) {
	query, err := os.ReadFile("/init.sql")
	if err != nil {
		log.Printf("file init.sql not found")
		return
	}

	_, err = db.Exec(string(query))
	if err != nil {
		log.Fatal("Failed to initialize db with SQL script:", err)
	}

	log.Println("The database has been initialized correctly")
}
