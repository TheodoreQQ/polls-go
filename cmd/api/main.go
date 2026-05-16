package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/TheodoreQQ/polls-go/docs"

	"github.com/TheodoreQQ/polls-go/internal/handlers"
	"github.com/TheodoreQQ/polls-go/internal/middleware"
	"github.com/TheodoreQQ/polls-go/internal/repository"
	"github.com/TheodoreQQ/polls-go/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Polls Project API
// @version         1.0
// @description     To jest serwer ankiet.
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Type 'Bearer ' followed by a space and then your token.
// @description             Wpisz: Bearer <twoj_token_jwt>
// @host      localhost:8080
// @BasePath  /
func main() {
	// Conntecting to database
	_ = godotenv.Load())
	// if err != nil {
	// 	log.Fatal("Failed to load .env file")
	// }
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		log.Fatal("URL is empty")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	defer db.Close()

	// Repo initialization
	pollRepo := &repository.PollRepository{DB: db}
	authRepo := &repository.AuthRepository{DB: db}

	hub := ws.NewHub()
	go hub.Run()

	// Handler initialization

	pollHandler := &handlers.PollHandler{
		Repo: pollRepo,
		Hub:  hub,
	}

	authHandler := &handlers.AuthHandler{
		Repo: authRepo,
	}

	// Gin router configuration
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Endpoints definitions
	r.Use(middleware.UsageLogger(db))
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.GET("/join/:code", pollHandler.GetPollByCode)
	r.POST("/vote", pollHandler.Vote)
	r.GET("/ws/:id", pollHandler.WSHandler)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.PersistAuthorization(true), ginSwagger.DocExpansion("list")))

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/polls", pollHandler.CreatePoll)
	protected.GET("/polls", pollHandler.GetPoll)
	protected.PATCH("/polls/:id/activate", pollHandler.ActivatePoll)
	protected.GET("/polls/:id/results", pollHandler.GetVotesByPoll)
	protected.DELETE("/polls/:id/delete", pollHandler.DeletePoll)
	protected.PATCH("/polls/:id/deactivate", pollHandler.DeactivatePoll)
	protected.PATCH("/polls/:id/question", pollHandler.UpdateQuestion)
	protected.GET("polls/:id/download", pollHandler.DownloadReport)

	// Start serwera
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
