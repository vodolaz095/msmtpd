package dovecot

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"msmtpd"
)

// Authenticate returns true if the password is valid for the user, false
// otherwise.
func (a *Auth) Authenticate(tr *msmtpd.Transaction, user, passwd string) error {
	if !isUsernameSafe(user) {
		tr.LogWarn("user %s is considered unsafe for dovecot usage", user)
		return permanentError
	}

	conn, err := a.dial("unix", a.PathToAuthClientSocket)
	if err != nil {
		tr.LogError(err, "while dialing address of dovecot's client socket")
		return temporaryError
	}
	defer conn.Close()
	tr.LogDebug("Dovecot responses seems sane on socket %s", a.PathToAuthClientSocket)

	// Send our version, and then our PID.
	err = write(conn, fmt.Sprintf("VERSION\t1\t1\nCPID\t%d\n", os.Getpid()))
	if err != nil {
		tr.LogError(err, "while receiving dovecot protocol version")
		return temporaryError
	}

	// Read the server-side handshake. We don't care about the contents
	// really, so just read all lines until we see the DONE.
	for {
		resp, readlineErr := conn.ReadLine()
		if readlineErr != nil {
			tr.LogError(err, "while receiving dovecot protocol version")
			return temporaryError
		}
		if resp == "DONE" {
			break
		}
	}

	// We only support PLAIN authentication, so construct the request.
	// Note we set the "secured" option, with the assumption that we got the
	// password via a secure channel (like TLS).
	// TODO: does dovecot handle utf8 domains well? do we need to encode them
	// in IDNA first?
	resp := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s\x00%s\x00%s", user, user, passwd)))

	err = write(conn, fmt.Sprintf(
		"AUTH\t1\tPLAIN\tservice=smtp\tsecured\tno-penalty\tnologin\tresp=%s\n", resp))
	if err != nil {
		tr.LogError(err, "while writing auth request to dovecot")
		return temporaryError
	}

	// Get the response, and we're done.
	resp, err = conn.ReadLine()
	if err != nil {
		tr.LogError(err, "while receiving dovecot authentication response")
		return temporaryError
	} else if strings.HasPrefix(resp, "OK\t1") {
		tr.LogInfo("Dovecot authorization passed for %s", user)
		return nil
	} else if strings.HasPrefix(resp, "FAIL\t1") {
		tr.LogInfo("Dovecot authorization failed for %s", user)
		return permanentError
	}
	tr.LogError(fmt.Errorf("invalid response: %q", resp), "while reading dovecot response for authentication")
	return temporaryError
}
