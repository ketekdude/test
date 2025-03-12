package errors

import (
	"errors"
)

type localError struct {
	error
	ErrorMessage string
}

type validationError struct {
	*localError
}

func NewInternalServerError(err error) error {
	return &localError{
		error: err,
	}
}

func NewValidationError(err string) error {
	return &validationError{
		&localError{
			error: errors.New(err),
		},
	}
}

func (le *localError) Error() string {
	if IsValidationError(le) {
		return "internal server error, please check log."
	}
	return le.error.Error()
}

func IsValidationError(err error) bool {
	_, ok := err.(*validationError)
	return ok
}
