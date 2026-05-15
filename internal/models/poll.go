package models

import "time"

type User struct {
	ID       int    `json:"id" swaggerignore:"true"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Poll struct {
	ID        int       `json:"id"`
	Question  string    `json:"question" binding:"required"`
	OwnerID   int       `json:"-"`
	Options   []Option  `json:"options" binding:"required"`
	Code      string    `json:"code"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Option struct {
	ID     int    `json:"id"`
	PollID int    `json:"poll_id"`
	Text   string `json:"text"`
	Votes  int    `json:"votes"`
}

type Vote struct {
	OptionID int `json:"option_id"`
	VoterID  int `json:"voter_id"`
}

type OptionResponse struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Votes int    `json:"votes"`
}
