package msmptd

import (
	"errors"
	"fmt"
)

// ErrorSMTP represents an Error reported in the SMTP session.
type ErrorSMTP struct {
	Code    int    // The integer error code
	Message string // The error message
}

// Error returns a string representation of the SMTP error
func (e ErrorSMTP) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

// ErrServerClosed is returned by the Server's Serve and ListenAndServe,
// methods after a call to shut down.
var ErrServerClosed = errors.New("smtp: Server closed")
