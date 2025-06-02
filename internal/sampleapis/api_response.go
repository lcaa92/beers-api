package sampleapis

type APIResponseError struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}
