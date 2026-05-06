package models

type ResponseForStudent struct {
	ID       int                 `json:"id"`
	Question string              `json:"question"`
	Options  []OptionsForStudent `json:"options"`
}

type OptionsForStudent struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}
