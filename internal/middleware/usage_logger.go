package middleware

import (
	"database/sql"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// function that logs information to the database
func UsageLogger(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// request goes to its handler
		c.Next()

		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		path := c.Request.URL.Path
		method := c.Request.Method

		userID, _ := c.Get("user_id")

		go func() {
			query := `
				INSERT INTO api_logs (user_id, action, status_code, path, latency_ms) 
				VALUES ($1, $2, $3, $4, $5)`

			_, err := db.Exec(query, userID, method, status, path, latency)
			if err != nil {
				log.Printf("Błąd podczas zapisu logów do bazy: %v", err)
			}
		}()
	}
}
