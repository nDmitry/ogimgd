package server

const statusError = "error"

// errorResponse is HTTP error request message format
type errorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// newErrorResponse returns an error response
func newErrorResponse(message string) errorResponse {
	return errorResponse{
		Status:  statusError,
		Message: message,
	}
}
