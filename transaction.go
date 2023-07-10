package msmptd

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"sync"
	"time"
)

// Protocol represents the protocol used in the SMTP session
type Protocol string

const (
	// SMTP means Simple Mail Transfer Protocol
	SMTP Protocol = "SMTP"

	// ESMTP means Extended Simple Mail Transfer Protocol, because it has some extra features
	// Simple Mail Transfer Protocol doesn't have
	ESMTP Protocol = "ESMTP"
)

// Transaction used to handle all SMTP protocol interactions with client
type Transaction struct {
	// ID is unique transaction identificator
	ID string `json:"id"`
	// StartedAt depicts moment when transaction was initiated
	StartedAt time.Time

	// ServerName depicts how out smtp server names itself
	ServerName string
	// Addr depicts network address of remote client
	Addr net.Addr
	// TLS Connection details, if encryption is enabled
	TLS *tls.ConnectionState
	// Encrypted means connection is encrypted by TLS
	Encrypted bool
	// Secured means TLS handshake succeeded
	Secured bool
	// HeloName is how client introduced himself via HELO/EHLO command
	HeloName string
	// Protocol used, SMTP or ESMTP
	Protocol Protocol
	// Username as provided by via authorization process command
	Username string
	// Password from authentication, if authenticated
	Password string
	// MailFrom stores address from which this message is originated as client says via `MAIL FROM:`
	MailFrom mail.Address
	// RcptTo stores addresses for which this message should be delivered as client says via `RCPT TO:`
	RcptTo []mail.Address

	// Body stores unparsed message body
	Body []byte

	// Parsed stores parsed message body
	Parsed *mail.Message
	// Logger is logging system inherited from server
	Logger Logger
	// facts are map of string data related to transaction
	facts map[string]string
	// counters are map of float data related to transaction
	counters map[string]float64
	// flags are map of bool data related to transaction
	flags map[string]bool

	ctx    context.Context
	cancel context.CancelFunc

	server  *Server
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	scanner *bufio.Scanner

	mu *sync.Mutex
}

// Context returns transaction context, which is canceled when transaction is closed
func (t *Transaction) Context() context.Context {
	return t.ctx
}

/*
 * Metadata manipulation
 */

// SetFact sets string parameter Transaction.facts
func (t *Transaction) SetFact(name, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.facts) == 0 {
		t.facts = make(map[string]string, 0)
	}
	t.facts[name] = value
}

// GetFact returns string fact from Transaction.facts
func (t *Transaction) GetFact(name string) (value string, found bool) {
	value, found = t.facts[name]
	return
}

// Incr increments transaction counter
func (t *Transaction) Incr(key string, delta float64) (newVal float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.facts) == 0 {
		t.counters = make(map[string]float64, 0)
	}
	old, found := t.counters[key]
	if found {
		newVal = old + delta
		t.counters[key] = newVal
		return newVal
	}
	t.counters[key] = delta
	return t.counters[key]
}

// GetCounter returns counter value
func (t *Transaction) GetCounter(key string) (val float64, found bool) {
	val, found = t.counters[key]
	return
}

// SetFlag set flag enabled for transaction
func (t *Transaction) SetFlag(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.flags[name] = true
}

// UnsetFlag unsets boolean flag from transaction
func (t *Transaction) UnsetFlag(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, found := t.flags[name]
	if found {
		delete(t.flags, name)
	}
}

// IsFlagSet returns true, if flag is set
func (t *Transaction) IsFlagSet(name string) bool {
	val, found := t.flags[name]
	if found {
		return val
	}
	return false
}

/*
 * Karma manipulation
 */

const karmaCounterName = "karma"

// Karma returns current transaction karma
func (t *Transaction) Karma() int {
	karma, found := t.counters[karmaCounterName]
	if found {
		return int(karma)
	}
	return 0
}

// Love grants good points to karma, promising message to enter Paradise for SMTP transactions, aka dovecot server socket for accepting messages via SMTP
func (t *Transaction) Love(delta int) (newVal int) {
	return int(t.Incr(karmaCounterName, float64(delta)))
}

// Hate grants bad points to karma, restricting message to enter Paradise for SMTP transactions, aka dovecot server socket for accepting messages via SMTP
func (t *Transaction) Hate(delta int) (newVal int) {
	return int(t.Incr(karmaCounterName, -float64(delta)))
}

/*
 * Header manipulation
 */

// AddHeader adds header, it should be called before AddReceivedLine, since it adds
// header to the top
func (t *Transaction) AddHeader(name, value string) {
	line := wrap([]byte(fmt.Sprintf("%s: %s\r\n", name, value)))
	t.Body = append(t.Body, line...)
	// Move the new newly added header line up front
	copy(t.Body[len(line):], t.Body[0:len(t.Body)-len(line)])
	copy(t.Body, line)
}

// AddReceivedLine prepends a Received header to the Data
func (t *Transaction) AddReceivedLine() {
	tlsDetails := ""
	tlsVersions := map[uint16]string{
		tls.VersionSSL30: "SSL3.0",
		tls.VersionTLS10: "TLS1.0",
		tls.VersionTLS11: "TLS1.1",
		tls.VersionTLS12: "TLS1.2",
		tls.VersionTLS13: "TLS1.3",
	}

	if t.TLS != nil {
		version := "unknown"
		if val, ok := tlsVersions[t.TLS.Version]; ok {
			version = val
		}
		cipher := tls.CipherSuiteName(t.TLS.CipherSuite)
		tlsDetails = fmt.Sprintf(
			"\r\n\t(version=%s cipher=%s);",
			version,
			cipher,
		)
	}
	peerIP := ""
	if addr, ok := t.Addr.(*net.TCPAddr); ok {
		peerIP = addr.IP.String()
	}
	line := wrap([]byte(fmt.Sprintf(
		"Received: from %s ([%s]) by %s with %s;%s\r\n\t%s\r\n",
		t.HeloName,
		peerIP,
		t.ServerName,
		t.Protocol,
		tlsDetails,
		time.Now().Format(timeFormatForHeaders),
	)))
	t.Body = append(t.Body, line...)
	// Move the new Received line up front
	copy(t.Body[len(line):], t.Body[0:len(t.Body)-len(line)])
	copy(t.Body, line)
}
