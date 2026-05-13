package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/TheodoreQQ/polls-go/internal/handlers"
	"github.com/TheodoreQQ/polls-go/internal/middleware"
	"github.com/TheodoreQQ/polls-go/internal/repository"
	"github.com/TheodoreQQ/polls-go/internal/ws"
	"github.com/gin-contrib/cors"
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
	pollRepo := &repository.PollRepository{DB: db}
	authRepo := &repository.AuthRepository{DB: db}

	hub := ws.NewHub()
	go hub.Run()

	pollHandler := &handlers.PollHandler{
		Repo: pollRepo,
		Hub:  hub,
	}

	authHandler := &handlers.AuthHandler{
		Repo: authRepo,
	}

	// Konfiguracja routera Gin
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Definicja endpointu
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.GET("/join/:code", pollHandler.GetPollByCode)
	r.POST("/vote", pollHandler.Vote)
	r.GET("/ws/:id", pollHandler.WSHandler)
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/polls", pollHandler.CreatePoll)
	protected.GET("/polls", pollHandler.GetPoll)
	protected.PATCH("/polls/:id/activate", pollHandler.ActivatePoll)
	protected.GET("/polls/:id/results", pollHandler.GetVotesByPoll)
	protected.DELETE("/polls/:id/delete", pollHandler.DeletePoll)
	protected.PATCH("/polls/:id/deactivate", pollHandler.DeactivatePoll)
	protected.PATCH("/polls/:id/question", pollHandler.UpdateQuestion)

	// Start serwera
	r.Run(":8080")
}
