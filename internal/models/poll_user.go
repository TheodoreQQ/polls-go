package models

import "time"

type ReponseForUser struct {
	ID        int       `json:"id"`
	Question  string    `json:"question"`
	Options   []Option  `json:"options"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type PollResultsResponse struct {
	ID         int            `json:"id"`
	Question   string         `json:"question"`
	Options    []OptionResult `json:"options"`
	TotalVotes int            `json:"total_votes"`
}

type OptionResult struct {
	ID         int    `json:"id"`
	Text       string `json:"text"`
	VotesCount int    `json:"votes"`
}

type UpdatePollWithOptionRequest struct {
	Question string `json:"question"`
	Options  []struct {
		ID   int    `json:"id"`
		Text string `json:"text"`
	} `json:"options"`
}
