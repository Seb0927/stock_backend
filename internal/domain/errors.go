package domain

import "errors"

var (
	// ErrNotFound indicates that the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates that the input provided is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicateEntry indicates that the entry already exists
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrDatabaseConnection indicates a database connection error
	ErrDatabaseConnection = errors.New("database connection error")

	// ErrExternalAPI indicates an error from external API
	ErrExternalAPI = errors.New("external API error")

	// ErrTimeout indicates a timeout error
	ErrTimeout = errors.New("operation timeout")

	// ErrUnauthorized indicates an unauthorized request
	ErrUnauthorized = errors.New("unauthorized")
)
