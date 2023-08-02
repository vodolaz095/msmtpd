package dovecot

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// CheckRecipient returns true if the user exists, false otherwise.
func (d *Dovecot) CheckRecipient(tr *msmtpd.Transaction, recipient *mail.Address) error {
	var user string
	alias, overrideFound := tr.GetFact(RecipientOverrideFact)
	if overrideFound {
		tr.LogDebug("Using dovecot alias %s for user %s", alias, recipient.String())
		user = alias
	} else {
		tr.LogDebug("Using dovecot alias %s for user %s", recipient.Address, recipient.String())
		user = recipient.Address
	}

	if !isUsernameSafe(user) {
		tr.LogWarn("user %s is considered unsafe for dovecot usage", user)
		return permanentError
	}

	conn, err := d.dial("unix", d.PathToAuthUserDBSocket)
	if err != nil {
		tr.LogError(err, "while getting dialing socket of dovecot's userdb")
		return temporaryError
	}
	defer conn.Close()
	tr.LogDebug("Dovecot connection established to socket %s", d.PathToAuthUserDBSocket)

	// Dovecot greets us with version and server pid.
	// VERSION\t<major>\t<minor>
	// SPID\t<pid>
	err = expect(conn, "VERSION\t1")
	if err != nil {
		tr.LogError(err, "while receiving dovecot protocol version")
		return temporaryError
	}
	tr.LogDebug("Dovecot responses seems sane on socket %s", d.PathToAuthUserDBSocket)
	err = expect(conn, "SPID\t")
	if err != nil {
		tr.LogError(err, "while receiving dovecot response with SPID")
		return temporaryError
	}

	// Send our version, and then the request.
	tr.LogDebug("Sending find user request to dovecot...")
	err = write(conn, "VERSION\t1\t1\n")
	if err != nil {
		tr.LogError(err, "while sending our protocol version to dovecot socket")
		return temporaryError
	}
	err = write(conn, fmt.Sprintf("USER\t1\t%s\tservice=smtp\n", user))
	if err != nil {
		tr.LogError(err, "while sending check user request on dovecot socket")
		return temporaryError
	}

	// Get the response, and we're done.
	resp, err := conn.ReadLine()
	if err != nil {
		tr.LogError(err, "while receiving error from dovecot")
		return temporaryError
	} else if strings.HasPrefix(resp, "USER\t1\t") {
		tr.LogInfo("Recipient %s is accepted by dovecot", recipient.String())
		return nil
	} else if strings.HasPrefix(resp, "NOTFOUND\t") {
		tr.LogInfo("Recipient %s is not accepted by dovecot", recipient.String())
		return permanentError
	}
	tr.LogError(fmt.Errorf("invalid response: %q", resp), "while reading dovecot response for checking recipient")
	return temporaryError
}
