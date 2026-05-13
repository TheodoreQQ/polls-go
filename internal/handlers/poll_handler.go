package handlers

import (
	// "fmt"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	// "time"

	"github.com/TheodoreQQ/polls-go/internal/models"
	"github.com/TheodoreQQ/polls-go/internal/repository"
	"github.com/TheodoreQQ/polls-go/internal/utils"
	"github.com/TheodoreQQ/polls-go/internal/ws"
	"github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	Repo *repository.PollRepository
	Hub  *ws.Hub
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	var p models.ReponseForUser

	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format"})
		return
	}

	userID, err := utils.GetUserId(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if p.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question cannot be empty"})
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

	err = h.Repo.Create(&p, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *PollHandler) GetPoll(c *gin.Context) {

	userID, err := utils.GetUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	polls, err := h.Repo.GetPollById(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch polls"})
		return
	}

	c.JSON(http.StatusOK, polls)
}

func (h *PollHandler) ActivatePoll(c *gin.Context) {
	strPollID := c.Param("id")
	pollID, err := strconv.Atoi(strPollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	userID, err := utils.GetUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	roomCode, err := h.Repo.Activate(pollID, userID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found or you don't have permission"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "poll activated succesfully",
		"code":    roomCode})
}

func (h *PollHandler) GetPollByCode(c *gin.Context) {
	code := c.Param("code")

	poll, err := h.Repo.GetPollByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (h *PollHandler) Vote(c *gin.Context) {
	fmt.Printf("DEBUG: Adres Huba: %p\n", h.Hub)
	var vote models.Vote
	if err := c.ShouldBindJSON(&vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve data"})
		return
	}

	pollID, err := h.Repo.GetPollIDByOption(vote.OptionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Option not found"})
		return
	}

	cookieName := fmt.Sprintf("voted_poll_%d", pollID)
	if _, err := c.Cookie(cookieName); err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You have already voted!"})
		return
	}

	actualPollID, err := h.Repo.VotePoll(vote.OptionID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Poll is deactivated or you chose wrong option"})
		return
	}

	results, err := h.Repo.GetResultsForBroadcast(actualPollID)
	fmt.Printf("Wysyłam do Huba: PollID=%d, Dane=%v\n", actualPollID, results) // DEBUG
	if err == nil {
		h.Hub.Broadcast <- ws.VoteUpdate{
			PollID: actualPollID,
			Data:   results,
		}
	} else {
		fmt.Printf("Broadcast Error: %v\n", err)
	}

	cookieName = fmt.Sprintf("voted_poll_%d", actualPollID)
	c.SetCookie(
		cookieName,
		"true",
		86400,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Vote successful"})
}

func (h *PollHandler) GetVotesByPoll(c *gin.Context) {
	strPollID := c.Param("id")
	pollID, err := strconv.Atoi(strPollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	userID, err := utils.GetUserId(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	response, err := h.Repo.GetVotesByPolll(pollID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found or you don't have permission"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PollHandler) DeletePoll(c *gin.Context) {
	strPollID := c.Param("id")
	pollID, err := strconv.Atoi(strPollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	userID, err := utils.GetUserId(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err = h.Repo.DeletePoll(pollID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found or no permission"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Poll deleted successfully"})
}

func (h *PollHandler) DeactivatePoll(c *gin.Context) {
	strPollID := c.Param("id")
	pollID, err := strconv.Atoi(strPollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	userID, err := utils.GetUserId(c)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retrieve data"})
		return
	}

	err = h.Repo.DeactivatePoll(pollID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found, no permission or is already deactivated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "poll status has been changed succesfully"})
}

func (h *PollHandler) UpdateQuestion(c *gin.Context) {
	strPollID := c.Param("id")
	pollID, err := strconv.Atoi(strPollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	userID, err := utils.GetUserId(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data"})
		return
	}

	var req models.UpdatePollWithOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question is empty"})
		return
	}

	err = h.Repo.UpdateQuestion(pollID, userID, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found or no permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *PollHandler) WSHandler(c *gin.Context) {
	pollID, _ := strconv.Atoi(c.Param("id"))

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.Hub.Mu.Lock()
	if h.Hub.Rooms == nil {
		h.Hub.Rooms = make(map[int]map[*websocket.Conn]bool)
	}
	if h.Hub.Rooms[pollID] == nil {
		h.Hub.Rooms[pollID] = make(map[*websocket.Conn]bool)
	}
	h.Hub.Rooms[pollID][conn] = true
	h.Hub.Mu.Unlock()

	defer func() {
		h.Hub.Mu.Lock()
		delete(h.Hub.Rooms[pollID], conn)
		h.Hub.Mu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
