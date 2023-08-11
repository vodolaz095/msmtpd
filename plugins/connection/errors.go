package connection

import "github.com/vodolaz095/msmtpd"

var friendlyError = msmtpd.ErrorSMTP{
	Code:    521,
	Message: "FUCK OFF!", // lol
}
