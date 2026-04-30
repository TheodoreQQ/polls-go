package models

import "time"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Poll struct {
	ID        int       `json:"id"`
	Question  string    `json:"question"`
	OwnerID   int       `json:"owner_id"`
	Options   []Option  `json:"options"`
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
