package deliver

import "github.com/vodolaz095/msmtpd"

var TemporaryError = msmtpd.ErrorSMTP{
	Code:    451,
	Message: "temporary errors, please, try again later",
}

var UnknownRecipientError = msmtpd.ErrorSMTP{
	Code:    521,
	Message: "i have no idea about recipient you want to deliver message to",
}
