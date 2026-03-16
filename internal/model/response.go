package models

type ResponsePayload struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
