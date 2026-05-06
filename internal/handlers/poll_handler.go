package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/TheodoreQQ/polls-go/internal/models"
	"github.com/TheodoreQQ/polls-go/internal/utils"
	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	DB *sql.DB
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	var p models.Poll

	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Question and options are required"})
		return
	}

	if len(p.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least two options are required"})
		return
	}

	for _, opt := range p.Options {
		if opt.Text == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Option text cannot be empty"})
			return
		}
	}

	userID, exists := c.Get("user_id")

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

func (h *PollHandler) ActivatePoll(c *gin.Context) {
	pollID := c.Param("id")
	roomCode := utils.GenerateCode()

	val, _ := c.Get("user_id")
	var userID int

	if id, ok := val.(int); ok {
		userID = id
	} else {
		userID = int(val.(float64))
	}

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE polls SET is_active = false, code = NULL WHERE owner_id = $1`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset polls"})
		return
	}

	result, err := tx.Exec(`UPDATE polls SET is_active = true, code = $1 WHERE id = $2 AND owner_id = $3`, roomCode, pollID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found or no permission"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "poll activated succesfully",
		"code":    roomCode})
}

func (h *PollHandler) GetPollByCode(c *gin.Context) {
	code := c.Param("code")

	query := `SELECT p.id, p.question, o.id, o.text 
						FROM polls p 
						JOIN options o ON p.id = o.poll_id 
						WHERE p.code = $1 AND p.is_active = true`

	rows, err := h.DB.Query(query, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data"})
		return
	}
	defer rows.Close()

	var poll models.ResponseForStudent
	first := true

	for rows.Next() {
		var pID, oID int
		var pQuestion, oText string

		if err := rows.Scan(&pID, &pQuestion, &oID, &oText); err != nil {
			continue
		}

		if first {
			poll.ID = pID
			poll.Question = pQuestion
			poll.Options = []models.OptionsForStudent{}
			first = false
		}

		poll.Options = append(poll.Options, models.OptionsForStudent{
			ID:   oID,
			Text: oText,
		})
	}
	if first {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active poll for that code"})
		return
	}
	c.JSON(http.StatusOK, poll)

}

func (h *PollHandler) Vote(c *gin.Context) {
	var vote models.Vote
	if err := c.ShouldBindJSON(&vote); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data"})
		return
	}

	var pollID int
	var isActive bool

	err := h.DB.QueryRow("SELECT p.id, p.is_active FROM polls p JOIN options o ON p.id = o.poll_id WHERE o.id = $1", vote.OptionID).Scan(&pollID, &isActive)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Option not found"})
		return
	}

	if !isActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Poll is not active"})
		return
	}
	cookieName := fmt.Sprintf("voted_poll_%d", pollID)
	if _, err := c.Cookie(cookieName); err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You have already voted!"})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}
	defer tx.Rollback()

	query := `UPDATE options SET votes_count = votes_count + 1 WHERE id = $1 AND poll_id IN (SELECT id FROM polls WHERE is_active = true) 
	RETURNING poll_id`

	err = tx.QueryRow(query, vote.OptionID).Scan(&pollID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to update votes"})
		return
	}

	cookieName = fmt.Sprintf("voted_poll_%d", pollID)
	c.SetCookie(
		cookieName,
		"true",
		86400,
		"/",
		"localhost",
		false,
		true,
	)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote successful"})
}

func (h *PollHandler) GetVotesByPoll(c *gin.Context) {
	pollID := c.Param("id")

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

	query := `SELECT p.id, p.question, o.id, o.text, o.votes_count FROM polls p JOIN options o ON p.id = o.poll_id WHERE p.id = $1 AND p.owner_id = $2`

	rows, err := h.DB.Query(query, pollID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data"})
		return
	}

	defer rows.Close()

	var response models.PollResultsResponse
	var totalVotes int

	for rows.Next() {
		var pID, oID, oVotes int
		var pQuestion, oText string

		err = rows.Scan(&pID, &pQuestion, &oID, &oText, &oVotes)
		if err != nil {
			continue
		}

		if response.ID == 0 {
			response.ID = pID
			response.Question = pQuestion
		}

		totalVotes += oVotes

		response.Options = append(response.Options, models.OptionResult{
			ID:         oID,
			Text:       oText,
			VotesCount: oVotes,
		})
	}

	if response.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found or no permission"})
		return
	}

	response.TotalVotes = totalVotes
	c.JSON(http.StatusOK, response)
}
