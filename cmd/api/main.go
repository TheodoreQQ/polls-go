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
	getHandler := &handlers.GetHandler{DB: db}
	authHandler := &handlers.AuthHandler{DB: db}

	// Konfiguracja routera Gin
	r := gin.Default()

	// Definicja endpointu
	r.POST("/polls", pollHandler.CreatePoll)
	r.GET("/options", getHandler.GetPoll)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Start serwera
	r.Run(":8080")
}
