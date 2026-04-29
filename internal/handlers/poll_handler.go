package handlers

import (
	"database/sql"
	"net/http"

	"github.com/TheodoreQQ/polls-go/internal/models"
	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	DB *sql.DB
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	var p models.Poll

	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queryPoll := `INSERT INTO polls (question) VALUES ($1) RETURNING id`
	err := h.DB.QueryRow(queryPoll, p.Question).Scan(&p.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Błąd zapisu do bazy"})
		return
	}

	for i := range p.Options {
		queryOption := `INSERT INTO options (poll_id, text) VALUES ($1, $2) RETURNING id`
		err := h.DB.QueryRow(queryOption, p.ID, p.Options[i].Text).Scan(&p.Options[i].ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Błąd zapisu opcji"})
			return
		}
		p.Options[i].PollID = p.ID
	}

	c.JSON(http.StatusCreated, p)
}
