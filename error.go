package msmtpd

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

// Status codes for SMTP negotiations
// see https://en.wikipedia.org/wiki/List_of_SMTP_server_return_codes

// ErrServiceNotAvailable means server cannot perform this SMTP transaction and it should be retried later
var ErrServiceNotAvailable = ErrorSMTP{
	Code:    421,
	Message: "Service not available, closing transmission channel. Try again later, please.",
}

// ErrServiceDoesNotAcceptEmail means server will not perform this SMTP transaction, even if your try to retry it
var ErrServiceDoesNotAcceptEmail = ErrorSMTP{
	Code:    521,
	Message: "Server does not accept mail. Do not retry delivery, please. It will fail.",
}

// ErrAuthenticationCredentialsInvalid means SMTP credentials are invalid
var ErrAuthenticationCredentialsInvalid = ErrorSMTP{
	Code:    535,
	Message: "Authentication credentials are invalid.",
}
