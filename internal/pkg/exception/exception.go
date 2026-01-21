package exception

import (
	"errors"
	"fmt"
)

// ApplicationError handles application level errors.
type ApplicationError struct {
	Message    string
	StatusCode int
	Cause      error
}

// Error interface implementation.
func (e ApplicationError) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %s", e.Message, e.Cause)
}

func (e ApplicationError) Unwrap() error {
	if e.Cause == nil {
		return errors.New(e.Message)
	}

	return e.Cause
}

func (e ApplicationError) Is(target error) bool {
	var targetErr ApplicationError

	if !errors.As(target, &targetErr) {
		return false
	}

	return e.Cause == targetErr.Cause &&
		e.Message == targetErr.Message
}

// ErrorCode returns error code for an application error.
func (e ApplicationError) ErrorCode() int {
	return e.StatusCode
}
