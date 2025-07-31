package errors

import (
	"fmt"
)

type ErrorType string

const (
	NetworkError    ErrorType = "network"
	ValidationError ErrorType = "validation"
	ConfigError     ErrorType = "config"
	PlayerError     ErrorType = "player"
	ScrapingError   ErrorType = "scraping"
)

type KaruError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *KaruError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *KaruError) Unwrap() error {
	return e.Cause
}

func New(errorType ErrorType, message string) *KaruError {
	return &KaruError{
		Type:    errorType,
		Message: message,
	}
}

func Wrap(err error, errorType ErrorType, message string) *KaruError {
	return &KaruError{
		Type:    errorType,
		Message: message,
		Cause:   err,
	}
}

func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *KaruError {
	return &KaruError{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}
