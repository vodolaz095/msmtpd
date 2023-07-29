package dovecot

import "github.com/vodolaz095/msmtpd"

// RecipientOverrideFact is name of fact to store username being used for dovecot
const RecipientOverrideFact = "DovecotRecipientOverrideFact"

// DefaultAuthUserSocketPath is path to dovecot socket being used for authorization
const DefaultAuthUserSocketPath = "/var/run/dovecot/auth-userdb"

// DefaultClientSocketPath is path to dovecot socket being used for checking if recipient exists
const DefaultClientSocketPath = "/var/run/dovecot/auth-client"

// DefaultLMTPSocketPath is path to dovecot socket which accepts email via LMTP protocol
const DefaultLMTPSocketPath = "/var/run/dovecot/lmtp"

var temporaryError = msmtpd.ErrorSMTP{
	Code:    451,
	Message: "temporary errors, please, try again later",
}

var permanentError = msmtpd.ErrorSMTP{
	Code:    521,
	Message: "i have no idea about recipient you want to deliver message to",
}
