package handlers

import (
	"database/sql"
	"net/http"

	"github.com/TheodoreQQ/polls-go/internal/models"
	"github.com/gin-gonic/gin"
)

type GetHandler struct {
	DB *sql.DB
}

func (h *GetHandler) GetPoll(c *gin.Context) {

	rows, err := h.DB.Query("SELECT id, question, is_active, created_at FROM polls")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	defer rows.Close()

	polls := []models.Poll{}

	for rows.Next() {
		var p models.Poll
		err := rows.Scan(&p.ID, &p.Question, &p.IsActive, &p.CreatedAt)
		if err != nil {
			continue
		}
		polls = append(polls, p)
	}

	c.IndentedJSON(http.StatusOK, polls)
}
