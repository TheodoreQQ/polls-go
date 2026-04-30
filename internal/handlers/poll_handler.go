package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/TheodoreQQ/polls-go/internal/models"
	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	DB *sql.DB
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	var p models.Poll
	userID, exists := c.Get("user_id")

	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if id, ok := userID.(int); ok {
		p.OwnerID = id
	} else if idFloat, ok := userID.(float64); ok {
		p.OwnerID = int(idFloat)
	}

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Błąd rozpoczęcia transakcji"})
		return
	}
	defer tx.Rollback()

	queryPoll := `INSERT INTO polls (question, owner_id) VALUES ($1, $2) RETURNING id`
	err = tx.QueryRow(queryPoll, p.Question, p.OwnerID).Scan(&p.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Błąd zapisu do bazy"})
		return
	}

	for i := range p.Options {
		queryOption := `INSERT INTO options (poll_id, text) VALUES ($1, $2) RETURNING id`
		err := tx.QueryRow(queryOption, p.ID, p.Options[i].Text).Scan(&p.Options[i].ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Błąd zapisu opcji"})
			return
		}
		p.Options[i].PollID = p.ID
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Błąd zatwierdzania transakcji"})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *PollHandler) GetPoll(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var userID int
	switch v := val.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Błędny format ID użytkownika"})
		return
	}

	queryPoll := `SELECT p.id, p.question, p.is_active, COALESCE(o.id, 0), COALESCE(o.text, ''), 
	COALESCE(o.votes_count, 0)
								FROM polls p 
								LEFT JOIN options o ON p.id = o.poll_id
								WHERE p.owner_id = $1
								ORDER BY p.id`

	rows, err := h.DB.Query(queryPoll, userID)
	if err != nil {
		fmt.Println("Błąd SQL", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Błąd pobierania"})
		return
	}

	defer rows.Close()

	pollsMap := make(map[int]*models.Poll)

	var pollsOrder []int

	for rows.Next() {
		var pID, oID, oVotes int
		var pQuestion, oText string
		var pActive bool

		err := rows.Scan(&pID, &pQuestion, &pActive, &oID, &oText, &oVotes)
		if err != nil {
			continue
		}

		if _, exists := pollsMap[pID]; !exists {
			pollsMap[pID] = &models.Poll{
				ID:       pID,
				Question: pQuestion,
				IsActive: pActive,
				Options:  []models.Option{},
			}
			pollsOrder = append(pollsOrder, pID)
		}
		if oID > 0 {
			pollsMap[pID].Options = append(pollsMap[pID].Options, models.Option{
				ID:     oID,
				PollID: pID,
				Text:   oText,
				Votes:  oVotes,
			})

		}
	}

	result := []models.Poll{}
	for _, pollID := range pollsOrder {
		result = append(result, *pollsMap[pollID])
	}
	c.JSON(http.StatusOK, result)
}
