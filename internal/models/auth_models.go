package models

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3" example:"TheodoreQQ"`
	Password string `json:"password" binding:"required,min=3" example:"123"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"TheodoreQQ"`
	Password string `json:"password" binding:"required" example:"123"`
}
