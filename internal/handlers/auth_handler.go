package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/TheodoreQQ/polls-go/internal/models"
	"github.com/TheodoreQQ/polls-go/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// intializing struct to handle HTTP requests related to user authentication
type AuthHandler struct {
	Repo *repository.AuthRepository
}

// register handler to create a new user account in DB
// @Summary      User Register
// @Tags         Authorization
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.RegisterRequest  true  "User Credentials"
// @Success      201          {object}  map[string]string
// @Failure		400 {object}	map[string]string "Status bad request"
// @Failure		409 {object}	map[string]string "Status conflict"
// @Failure		500 {object}	map[string]string "Internal server error"
// @Router       /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error:" + " Username must be at least 3 characters long  as well as password"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	err = h.Repo.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})

}

// login handles user authentication and return JWT token
// @Summary      User Login
// @Tags         Authorization
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.LoginRequest  true  "User Credentials"
// @Success      200          {object}  map[string]string
// @Failure		400 {object}	models.ErrorBadRequest "Status unauthorized"
// @Failure		401 {object}	models.ErrorUnauthorized "Status unauthorized"
// @Failure		500 {object}	models.ErrorInternalServer "Internal server error"
// @Router       /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var u models.LoginRequest

	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "username and password are required"})
		return
	}

	storedID, storedHash, err := h.Repo.GetUser(u.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(u.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": storedID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
