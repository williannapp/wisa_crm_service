package errors

import (
	"encoding/json"
	"fmt"
)

// AppError represents a standardized application error for HTTP responses.
// It contains a code, user-friendly message, optional details, and HTTP status.
// Details must never include stack traces or internal error messages.
type AppError struct {
	Status    string `json:"status"`
	Code      string `json:"error_code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	HTTPStatus int   `json:"-"`
}

// ErrorResponse is the JSON-serializable structure for API responses.
// AppError is used as the internal type; this struct ensures correct JSON output.
type errorResponse struct {
	Status  string `json:"status"`
	Code    string `json:"error_code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewAppError creates a new AppError with validation.
// message must not be empty; httpStatus must be between 400 and 599.
// Details are optional and will be omitted from JSON when empty.
func NewAppError(code, message, details string, httpStatus int) *AppError {
	if message == "" {
		message = "Erro interno. Tente novamente."
	}
	if httpStatus < 400 || httpStatus > 599 {
		httpStatus = 500
	}
	return &AppError{
		Status:     "error",
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
	}
}

// Error implements the error interface for compatibility with code that treats errors.
// Returns a format suitable for logs; never use this output for client responses.
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s — %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// MarshalJSON customizes JSON serialization to include status and use snake_case.
func (e *AppError) MarshalJSON() ([]byte, error) {
	resp := errorResponse{
		Status:  e.Status,
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
	return json.Marshal(resp)
}
