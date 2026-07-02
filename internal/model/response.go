package models

// ResponsePayload represents a generic API response with status and optional message.
// generate:reset
type ResponsePayload struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message,omitempty" example:"successfully saved 5 metrics"`
}
