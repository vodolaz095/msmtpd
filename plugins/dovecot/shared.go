// Package dovecot implements functions to interact with Dovecot's
// authentication service.
//
// https://wiki.dovecot.org/Design/AuthProtocol
// https://wiki.dovecot.org/Services#auth
// Code is partially borrowed from https://github.com/albertito/chasquid/tree/master/internal/dovecot

package dovecot

import (
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"
	"unicode"
)

// DefaultTimeout to use. We expect Dovecot to be quite fast, but don't want
// to hang forever if something gets stuck.
const DefaultTimeout = 5 * time.Second

// Dovecot represents a particular Dovecot auth service to use.
type Dovecot struct {
	// PathToAuthUserDBSocket is path for dovecot socket being used in Exists command to check if recipient exists
	PathToAuthUserDBSocket string
	// PathToAuthClientSocket is path for dovecot socket being used in Authenticate command to check if sender
	// provided correct username and password
	PathToAuthClientSocket string

	// LtmpSocket is LMTP protocol socket for dovecot to accept email for local delivery
	LtmpSocket string

	// Timeout for connection and I/O operations (applies on each call).
	// Set to DefaultTimeout by NewAuth.
	Timeout time.Duration
}

func (d *Dovecot) dial(network, addr string) (*textproto.Conn, error) {
	nc, err := net.DialTimeout(network, addr, d.Timeout)
	if err != nil {
		return nil, err
	}
	err = nc.SetDeadline(time.Now().Add(d.Timeout))
	if err != nil {
		return nil, err
	}
	return textproto.NewConn(nc), nil
}

func expect(conn *textproto.Conn, prefix string) error {
	resp, err := conn.ReadLine()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(resp, prefix) {
		return fmt.Errorf("got %q", resp)
	}
	return nil
}

func write(conn *textproto.Conn, msg string) error {
	_, err := conn.W.Write([]byte(msg))
	if err != nil {
		return err
	}
	return conn.W.Flush()
}

// isUsernameSafe to use in the dovecot protocol?
// Unfotunately dovecot's protocol is not very robust wrt. whitespace,
// so we need to be careful.
func isUsernameSafe(user string) bool {
	for _, r := range user {
		if unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
