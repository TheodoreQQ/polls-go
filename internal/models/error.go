package models

type APIError struct {
	Error string `json:"error" example:"An unexpected error occured"`
}

// 401
type ErrorUnauthorized struct {
	Error string `json:"error" example:"Status unauthorized"`
}

// 404
type ErrorNotFound struct {
	Error string `json:"error" example:"Poll not found"`
}

// 400
type ErrorBadRequest struct {
	Error string `json:"error" example:"Invalid input"`
}

// 403
type ErrorForbiddenVote struct {
	Error string `json:"error" example:"You have already voted"`
}

// 500
type ErrorInternalServer struct {
	Error string `json:"error" example:"Internal server error"`
}
