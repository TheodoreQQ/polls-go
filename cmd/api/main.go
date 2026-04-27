package main

import (
	"database/sql"
	"log"

	"github.com/TheodoreQQ/polls-go/internal/handlers"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Połączenie z bazą
	connStr := "host=localhost port=5432 user=user password=password dbname=polls_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Inicjalizacja handlera
	pollHandler := &handlers.PollHandler{DB: db}

	// Konfiguracja routera Gin
	r := gin.Default()

	// Definicja endpointu
	r.POST("/polls", pollHandler.CreatePoll)

	// Start serwera
	r.Run(":8080")
}
