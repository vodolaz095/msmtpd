package deliver

import "github.com/vodolaz095/msmtpd"

// TemporaryError means backend is malfunctioning, but you can try to deliver later
var TemporaryError = msmtpd.ErrorSMTP{
	Code:    451,
	Message: "temporary errors, please, try again later",
}

// UnknownRecipientError means backend is unaware of recipient you want to deliver too
var UnknownRecipientError = msmtpd.ErrorSMTP{
	Code:    521,
	Message: "i have no idea about recipient you want to deliver message to",
}
