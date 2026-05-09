package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func GetUserId(c *gin.Context) (int, error) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, nil
	}

	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, errors.New("incorrect user_id format")
	}
}
