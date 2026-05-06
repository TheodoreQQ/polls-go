package models

type ReponseForUser struct {
	ID       int      `json:"id"`
	Question string   `json:"question"`
	Options  []Option `json:"options"`
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
