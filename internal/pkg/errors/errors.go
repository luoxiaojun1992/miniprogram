package errors

import "net/http"

// AppError represents a structured application error.
type AppError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	HTTPCode int    `json:"-"`
	Cause    error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewBadRequest creates a 400 bad request error.
func NewBadRequest(message string, cause error) *AppError {
	return &AppError{Code: 400001, Message: message, HTTPCode: http.StatusBadRequest, Cause: cause}
}

// NewValidation creates a 400 validation error.
func NewValidation(message string, cause error) *AppError {
	return &AppError{Code: 400002, Message: message, HTTPCode: http.StatusBadRequest, Cause: cause}
}

// NewUnauthorized creates a 401 unauthorized error.
func NewUnauthorized(message string, cause error) *AppError {
	return &AppError{Code: 401001, Message: message, HTTPCode: http.StatusUnauthorized, Cause: cause}
}

// NewForbidden creates a 403 forbidden error.
func NewForbidden(message string, cause error) *AppError {
	return &AppError{Code: 403001, Message: message, HTTPCode: http.StatusForbidden, Cause: cause}
}

// NewNotFound creates a 404 not found error.
func NewNotFound(message string, cause error) *AppError {
	return &AppError{Code: 404001, Message: message, HTTPCode: http.StatusNotFound, Cause: cause}
}

// NewConflict creates a 409 conflict error.
func NewConflict(message string, cause error) *AppError {
	return &AppError{Code: 409001, Message: message, HTTPCode: http.StatusConflict, Cause: cause}
}

// NewInternal creates a 500 internal error.
func NewInternal(message string, cause error) *AppError {
	return &AppError{Code: 500001, Message: message, HTTPCode: http.StatusInternalServerError, Cause: cause}
}

// ToResponse converts an error to an HTTP response tuple.
func ToResponse(err error) (int, map[string]interface{}) {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPCode, map[string]interface{}{
			"code":    appErr.Code,
			"message": appErr.Message,
			"data":    nil,
		}
	}
	return http.StatusInternalServerError, map[string]interface{}{
		"code":    500000,
		"message": "服务器内部错误",
		"data":    nil,
	}
}
