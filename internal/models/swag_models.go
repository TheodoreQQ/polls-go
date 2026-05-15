package models

type CreatePollRequest struct {
	Question string                `json:"question" example:"Be or not to be?"`
	Options  []CreateOptionRequest `json:"options"`
}

type CreateOptionRequest struct {
	Text string `json:"text" example:"Not to be"`
}

type VoteRequest struct {
	OptionID int `json:"option_id" example:"35"`
}

type UpdatePollRequest struct {
	Question string `json:"question" example:"Changed question"`
}
