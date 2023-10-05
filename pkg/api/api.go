package api

// SingleRequest is the request body for a single request
type SingleRequest map[string]interface{}

// BatchResponse is the response body for a batch request
type BatchResponse struct {
	Code int    `json:"code"`
	Body string `json:"body"`
}

// HTTPErrorResponse represents the typical API error response body
type HTTPErrorResponse struct {
	Error HTTPErrorContent `json:"error"` // Error content json object
}

// HTTPErrorContent holds all the relevant information regarding an API error
type HTTPErrorContent struct {
	Message string `json:"message"` // Human readable error message
	Code    int    `json:"code"`    // Error code
	ID      string `json:"id"`      // Trace ID of error to cross-reference with logs
}

// MakeHTTPErrorResponse makes and returns a populated HTTPErrorResponse struct
func MakeHTTPErrorResponse(msg string, code int, trackID string) HTTPErrorResponse {
	return HTTPErrorResponse{HTTPErrorContent{
		Message: msg,
		Code:    code,
		ID:      trackID,
	}}
}
