package errors

import (
	"errors"
	"io"
	"os"

	"github.com/cloudbridgeuy/scripts/pkg/logger"
	"github.com/cloudbridgeuy/scripts/pkg/term"
)

// HandleErrorWithReason logs an error message and returns an error.
func HandleErrorWithReason(err error, reason string) {
	HandleError(NewReasonError(err, reason))
}

// HandleError logs an error message and returns an error.
func HandleError(err error) {
	// exhaust stdin
	if !term.IsInputTTY() {
		_, _ = io.ReadAll(os.Stdin)
	}

	format := "\n%s\n\n%s\n\n"

	var args []interface{}
	var perr ReasonError

	if errors.As(err, &perr) {
		args = []interface{}{
			term.StderrStyles().ErrPadding.Render(term.StderrStyles().ErrorHeader.String(), perr.Reason()),
			term.StderrStyles().ErrPadding.Render(term.StderrStyles().ErrorDetails.Render(perr.Error())),
		}
	} else {
		args = []interface{}{
			term.StderrStyles().ErrPadding.Render(term.StderrStyles().ErrorDetails.Render(err.Error())),
		}
	}

	logger.Logger.Printf(format, args...)

	os.Exit(1)
}

// ReasonError is a wrapper around an error that adds additional context.
type ReasonError struct {
	err    error
	reason string
}

// NewReasonError creates a new scriptsError.
func NewReasonError(err error, reason string) ReasonError {
	return ReasonError{err, reason}
}

// Error returns the error message.
func (m ReasonError) Error() string {
	return m.err.Error()
}

// Reason returns the reason for the error.
func (m ReasonError) Reason() string {
	return m.reason
}
