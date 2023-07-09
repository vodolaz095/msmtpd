package msmptd

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIFkzCCA3ugAwIBAgIUQvhoyGmvPHq8q6BHrygu4dPp0CkwDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDESMBAGA1UEAwwJbG9jYWxob3N0MB4X
DTIwMDUyMTE2MzI1NVoXDTMwMDUxOTE2MzI1NVowWTELMAkGA1UEBhMCQVUxEzAR
BgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5
IEx0ZDESMBAGA1UEAwwJbG9jYWxob3N0MIICIjANBgkqhkiG9w0BAQEFAAOCAg8A
MIICCgKCAgEAk773plyfK4u2uIIZ6H7vEnTb5qJT6R/KCY9yniRvCFV+jCrISAs9
0pgU+/P8iePnZRGbRCGGt1B+1/JAVLIYFZuawILHNs4yWKAwh0uNpR1Pec8v7vpq
NpdUzXKQKIqFynSkcLA8c2DOZwuhwVc8rZw50yY3r4i4Vxf0AARGXapnBfy6WerR
/6xT7y/OcK8+8aOirDQ9P6WlvZ0ynZKi5q2o1eEVypT2us9r+HsCYosKEEAnjzjJ
wP5rvredxUqb7OupIkgA4Nq80+4tqGGQfWetmoi3zXRhKpijKjgxBOYEqSUWm9ws
/aC91Iy5RawyTB0W064z75OgfuI5GwFUbyLD0YVN4DLSAI79GUfvc8NeLEXpQvYq
+f8P+O1Hbv2AQ28IdbyQrNefB+/WgjeTvXLploNlUihVhpmLpptqnauw/DY5Ix51
w60lHIZ6esNOmMQB+/z/IY5gpmuo66yH8aSCPSYBFxQebB7NMqYGOS9nXx62/Bn1
OUVXtdtrhfbbdQW6zMZjka0t8m83fnGw3ISyBK2NNnSzOgycu0ChsW6sk7lKyeWa
85eJGsQWIhkOeF9v9GAIH/qsrgVpToVC9Krbk+/gqYIYF330tHQrzp6M6LiG5OY1
P7grUBovN2ZFt10B97HxWKa2f/8t9sfHZuKbfLSFbDsyI2JyNDh+Vk0CAwEAAaNT
MFEwHQYDVR0OBBYEFOLdIQUr3gDQF5YBor75mlnCdKngMB8GA1UdIwQYMBaAFOLd
IQUr3gDQF5YBor75mlnCdKngMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
BQADggIBAGddhQMVMZ14TY7bU8CMuc9IrXUwxp59QfqpcXCA2pHc2VOWkylv2dH7
ta6KooPMKwJ61d+coYPK1zMUvNHHJCYVpVK0r+IGzs8mzg91JJpX2gV5moJqNXvd
Fy6heQJuAvzbb0Tfsv8KN7U8zg/ovpS7MbY+8mRJTQINn2pCzt2y2C7EftLK36x0
KeBWqyXofBJoMy03VfCRqQlWK7VPqxluAbkH+bzji1g/BTkoCKzOitAbjS5lT3sk
oCrF9N6AcjpFOH2ZZmTO4cZ6TSWfrb/9OWFXl0TNR9+x5c/bUEKoGeSMV1YT1SlK
TNFMUlq0sPRgaITotRdcptc045M6KF777QVbrYm/VH1T3pwPGYu2kUdYHcteyX9P
8aRG4xsPGQ6DD7YjBFsif2fxlR3nQ+J/l/+eXHO4C+eRbxi15Z2NjwVjYpxZlUOq
HD96v516JkMJ63awbY+HkYdEUBKqR55tzcvNWnnfiboVmIecjAjoV4zStwDIti9u
14IgdqqAbnx0ALbUWnvfFloLdCzPPQhgLHpTeRSEDPljJWX8rmy8iQtRb0FWYQ3z
A2wsUyutzK19nt4hjVrTX0At9ku3gMmViXFlbvyA1Y4TuhdUYqJauMBrWKl2ybDW
yhdKg/V3yTwgBUtb3QO4m1khNQjQLuPFVxULGEA38Y5dXSONsYnt
-----END CERTIFICATE-----`)

var localhostKey = []byte(`-----BEGIN PRIVATE KEY-----
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
hBgHM6A0WJC9MO3aAKRBcp48y6DXNA==
-----END PRIVATE KEY-----`)

func cmd(c *textproto.Conn, expectedCode int, format string, args ...interface{}) error {
	id, err := c.Cmd(format, args...)
	if err != nil {
		return err
	}
	c.StartResponse(id)
	_, _, err = c.ReadResponse(expectedCode)
	c.EndResponse(id)
	return err
}

type testLogger struct{}

func (tl *testLogger) Tracef(transaction *Transaction, format string, args ...any) {
	fmt.Printf("TRACE: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Debugf(transaction *Transaction, format string, args ...any) {
	fmt.Printf("DEBUG: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Infof(transaction *Transaction, format string, args ...any) {
	fmt.Printf("INFO: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Warnf(transaction *Transaction, format string, args ...any) {
	fmt.Printf("WARN: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Errorf(transaction *Transaction, format string, args ...any) {
	fmt.Printf("ERROR: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Fatalf(transaction *Transaction, format string, args ...any) {
	panic("it is bad")
}

func AuthenticatorForTestsThatAlwaysWorks(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and succeed!", username, password)
	return nil
}

func AuthenticatorForTestsThatAlwaysFails(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and fail!", username, password)
	return ErrorSMTP{Code: 550, Message: "Denied"}
}

func runserver(t *testing.T, server *Server) (addr string, closer func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("Listen failed: %v", err)
	}
	logger := testLogger{}
	server.Logger = &logger
	go func() {
		serveErr := server.Serve(ln)
		if err != nil {
			t.Errorf("%s : while starting server on %s",
				serveErr, server.Address())
		}
	}()
	done := make(chan bool)
	go func() {
		<-done
		ln.Close()
	}()
	return ln.Addr().String(), func() {
		done <- true
	}
}

func runsslserver(t *testing.T, server *Server) (addr string, closer func()) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Errorf("Cert load failed: %v", err)
	}
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return runserver(t, server)
}

func TestSMTP(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if supported, _ := c.Extension("AUTH"); supported {
		t.Error("AUTH supported before TLS")
	}
	if supported, _ := c.Extension("8BITMIME"); !supported {
		t.Error("8BITMIME not supported")
	}
	if supported, _ := c.Extension("STARTTLS"); supported {
		t.Error("STARTTLS supported")
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	err = c.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}

	err = c.Verify("foobar@example.net")
	if err == nil {
		t.Error("Unexpected support for VRFY")
	}

	if err = cmd(c.Text, 250, "NOOP"); err != nil {
		t.Errorf("NOOP failed: %v", err)
	}

	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestListenAndServe(t *testing.T) {
	server := &Server{}
	addr, closer := runserver(t, server)
	closer()
	go func() {
		lsErr := server.ListenAndServe(addr)
		if lsErr != nil {
			t.Errorf("%s : while starting server on %s", lsErr, server.Address())
		}
	}()
	time.Sleep(100 * time.Millisecond)
	if server.Address().String() != addr {
		t.Errorf("server is listening on `%s` instead of `%s",
			server.Address(), addr,
		)
	}
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestSTARTTLS(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysWorks,
		ForceTLS:      true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if supported, _ := c.Extension("AUTH"); supported {
		t.Error("AUTH supported before TLS")
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Error("Mail workded before TLS with ForceTLS")
	}
	if err = cmd(c.Text, 220, "STARTTLS"); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = cmd(c.Text, 250, "foobar"); err == nil {
		t.Error("STARTTLS didn't fail with invalid handshake")
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err == nil {
		t.Error("STARTTLS worked twice")
	}
	if supported, _ := c.Extension("AUTH"); !supported {
		t.Error("AUTH not supported after TLS")
	}
	if _, mechs := c.Extension("AUTH"); !strings.Contains(mechs, "PLAIN") {
		t.Error("PLAIN AUTH not supported after TLS")
	}
	if _, mechs := c.Extension("AUTH"); !strings.Contains(mechs, "LOGIN") {
		t.Error("LOGIN AUTH not supported after TLS")
	}
	if err = c.Auth(smtp.PlainAuth("foo", "foo", "bar", "127.0.0.1")); err != nil {
		t.Errorf("Auth failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestAuthRejection(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysFails,
		ForceTLS:      true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Auth(smtp.PlainAuth("foo", "foo", "bar", "127.0.0.1")); err == nil {
		t.Error("Auth worked despite rejection")
	}
}

func TestAuthNotSupported(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		ForceTLS: true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Auth(smtp.PlainAuth("foo", "foo", "bar", "127.0.0.1")); err == nil {
		t.Error("Auth worked despite no authenticator")
	}
}

func TestAuthBypass(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysFails,
		ForceTLS:      true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Error("Unexpected MAIL success")
	}
}

func TestConnectionCheck(t *testing.T) {
	cc := make([]func(tr *Transaction) error, 0)
	cc = append(cc, func(tr *Transaction) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := runserver(t, &Server{
		ConnectionCheckers: cc,
	})
	defer closer()
	if _, err := smtp.Dial(addr); err == nil {
		t.Error("Dial succeeded despite ConnectionCheck")
	}
}

func TestConnectionCheckSimpleError(t *testing.T) {
	cc := make([]func(tr *Transaction) error, 0)
	cc = append(cc, func(tr *Transaction) error {
		return errors.New("Denied")
	})
	addr, closer := runserver(t, &Server{
		ConnectionCheckers: cc,
	})
	defer closer()
	if _, err := smtp.Dial(addr); err == nil {
		t.Error("Dial succeeded despite ConnectionCheck")
	}
}

func TestHELOCheck(t *testing.T) {
	addr, closer := runserver(t, &Server{
		HeloCheckers: []CheckerFunc{
			func(transaction *Transaction, name string) error {
				if name != "foobar.local" {
					t.Error("Wrong HELO name")
				}
				return ErrorSMTP{Code: 552, Message: "Denied"}
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("foobar.local"); err == nil {
		t.Error("Unexpected HELO success")
	}
}

func TestSenderCheck(t *testing.T) {
	sc := make([]CheckerFunc, 0)
	sc = append(sc, func(tr *Transaction, name string) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := runserver(t, &Server{
		SenderCheckers: sc,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Error("Unexpected MAIL success")
	}
}

func TestRecipientCheck(t *testing.T) {
	rc := make([]CheckerFunc, 0)
	rc = append(rc, func(tr *Transaction, name string) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := runserver(t, &Server{
		RecipientCheckers: rc,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err == nil {
		t.Error("Unexpected RCPT success")
	}
}

func TestMaxMessageSize(t *testing.T) {
	addr, closer := runserver(t, &Server{
		MaxMessageSize: 5,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err == nil {
		t.Error("Allowed message larger than 5 bytes to pass.")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestHandler(t *testing.T) {
	handlers := make([]func(tr *Transaction) error, 0)
	handlers = append(handlers, func(tr *Transaction) error {
		if tr.MailFrom.Address != "sender@example.org" {
			t.Errorf("Unknown sender: %v", tr.MailFrom)
		}
		if len(tr.RcptTo) != 1 {
			t.Errorf("Too many recipients: %d", len(tr.RcptTo))
		}
		if tr.RcptTo[0].Address != "recipient@example.net" {
			t.Errorf("Unknown recipient: %v", tr.RcptTo[0].Address)
		}
		if string(tr.Body) != "This is the email body\n" {
			t.Errorf("Wrong message body: %v", string(tr.Body))
		}
		return nil
	})
	addr, closer := runserver(t, &Server{
		Handlers: handlers,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestRejectHandler(t *testing.T) {
	handlers := make([]func(tr *Transaction) error, 0)
	handlers = append(handlers, func(tr *Transaction) error {
		return ErrorSMTP{Code: 550, Message: "Rejected"}
	})
	addr, closer := runserver(t, &Server{
		Handlers: handlers,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err == nil {
		t.Error("Unexpected accept of data")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestMaxConnections(t *testing.T) {
	addr, closer := runserver(t, &Server{
		MaxConnections: 1,
	})
	defer closer()
	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	_, err = smtp.Dial(addr)
	if err == nil {
		t.Error("Dial succeeded despite MaxConnections = 1")
	}
	c1.Close()
}

func TestNoMaxConnections(t *testing.T) {
	addr, closer := runserver(t, &Server{
		MaxConnections: -1,
	})
	defer closer()
	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	c1.Close()
}

func TestMaxRecipients(t *testing.T) {
	addr, closer := runserver(t, &Server{
		MaxRecipients: 1,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err == nil {
		t.Error("RCPT succeeded despite MaxRecipients = 1")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestInvalidHelo(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello(""); err == nil {
		t.Error("Unexpected HELO success")
	}
}

func TestInvalidSender(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("invalid@@example.org"); err == nil {
		t.Error("Unexpected MAIL success")
	}
}

func TestInvalidRecipient(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("invalid@@example.org"); err == nil {
		t.Error("Unexpected RCPT success")
	}
}

func TestRCPTbeforeMAIL(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err == nil {
		t.Error("Unexpected RCPT success")
	}
}

func TestDATAbeforeRCPT(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if _, err = c.Data(); err == nil {
		t.Error("Data accepted despite no recipients")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestInterruptedDATA(t *testing.T) {
	handlers := make([]func(tr *Transaction) error, 0)
	handlers = append(handlers, func(tr *Transaction) error {
		t.Error("Accepted DATA despite disconnection")
		return nil
	})
	addr, closer := runserver(t, &Server{
		Handlers: handlers,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	c.Close()
}

func TestMeta(t *testing.T) {
	addr, closer := runserver(t, &Server{
		MaxConnections: 1,
		HeloCheckers: []CheckerFunc{
			func(transaction *Transaction, name string) error {
				transaction.SetFact("something", name)
				transaction.Incr("int64", 1)
				transaction.Incr("float64", 1.1)
				transaction.LogWarn("something")
				return nil
			},
		},
		SenderCheckers: []CheckerFunc{
			func(transaction *Transaction, name string) error {
				var found bool
				_, found = transaction.GetFact("nothing")
				if found {
					t.Error("fact `nothing` is found?")
				}
				something, found := transaction.GetFact("something")
				if !found {
					t.Errorf("fact `something` is not set!")
				}
				if something != "localhost" {
					t.Errorf("wrong meta `something` %s instead of `localhost`", something)
				}
				integerValue, found := transaction.GetCounter("int64")
				if !found {
					t.Errorf("counter `int64` is not set!")
				}
				if integerValue != 1 {
					t.Errorf("wrong value for `int64`")
				}
				floatValue, found := transaction.GetCounter("float64")
				if !found {
					t.Errorf("counter `float64` is not set!")
				}
				if floatValue != 1.1 {
					t.Errorf("wrong value for `float64` - %v", floatValue)
				}
				_, found = transaction.GetCounter("lalala")
				if found {
					t.Errorf("unexistend counter returned value")
				}
				transaction.Incr("int64", 1)
				transaction.Incr("float64", 1.1)
				return nil
			},
		},
		RecipientCheckers: []CheckerFunc{
			func(transaction *Transaction, name string) error {
				var found bool
				a, found := transaction.GetCounter("int64")
				if !found {
					t.Errorf("counter `int64` is not set!")
				}
				b, found := transaction.GetCounter("float64")
				if !found {
					t.Errorf("counter `float64` is not set!")
				}
				c, found := transaction.GetFact("something")
				if !found {
					t.Errorf("fact `something` is not set!")
				}
				return ErrorSMTP{
					Code:    451,
					Message: fmt.Sprintf("%v %v %s", a, b, c),
				}
			},
		},
	})

	defer closer()
	cm, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = cm.Hello("localhost")
	if err != nil {
		t.Error(err)
	}
	err = cm.Mail("somebody@localhost")
	if err != nil {
		t.Error(err)
	}
	err = cm.Rcpt("scuba@example.org")
	if err != nil {
		if err.Error() != "451 2 2.2 localhost" {
			t.Errorf("wrong error `%s` instead `451 2 2.2 localhost`", err)
		}
	}
	err = cm.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestTimeoutClose(t *testing.T) {
	addr, closer := runserver(t, &Server{
		MaxConnections: 1,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	})
	defer closer()
	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	time.Sleep(time.Second * 2)
	c2, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c1.Mail("sender@example.org"); err == nil {
		t.Error("MAIL succeeded despite being timed out.")
	}
	if err = c2.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c2.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
	c2.Close()
}

func TestTLSTimeout(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		ReadTimeout:  time.Second * 2,
		WriteTimeout: time.Second * 2,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestLongLine(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail(fmt.Sprintf("%s@example.org", strings.Repeat("x", 65*1024))); err == nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestXCLIENT(t *testing.T) {
	sc := make([]CheckerFunc, 0)
	sc = append(sc, func(tr *Transaction, name string) error {
		if tr.HeloName != "new.example.net" {
			t.Errorf("Didn't override HELO name: %v", tr.HeloName)
		}
		if tr.Addr.String() != "42.42.42.42:4242" {
			t.Errorf("Didn't override IP/Port: %v", tr.Addr)
		}
		if tr.Username != "newusername" {
			t.Errorf("Didn't override username: %v", tr.Username)
		}
		if tr.Protocol != SMTP {
			t.Errorf("Didn't override protocol: %v", tr.Protocol)
		}
		return nil
	})
	addr, closer := runserver(t, &Server{
		EnableXCLIENT:  true,
		SenderCheckers: sc,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if supported, _ := c.Extension("XCLIENT"); !supported {
		t.Error("XCLIENT not supported")
	}
	err = cmd(c.Text, 220, "XCLIENT NAME=ignored ADDR=42.42.42.42 PORT=4242 PROTO=SMTP HELO=new.example.net LOGIN=newusername")
	if err != nil {
		t.Errorf("XCLIENT failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestEnvelopeReceived(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		Hostname: "foobar.example.net",
		Handlers: []func(tr *Transaction) error{
			func(tr *Transaction) error {
				tr.AddReceivedLine()
				if !bytes.HasPrefix(tr.Body, []byte("Received: from localhost ([127.0.0.1]) by foobar.example.net with ESMTP;")) {
					t.Error("Wrong received line.")
				}
				return nil
			},
		},
		ForceTLS: true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestExtraHeader(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		Hostname: "foobar.example.net",
		Handlers: []func(tr *Transaction) error{
			func(tr *Transaction) error {
				tr.AddHeader("Something", "interesting")
				if !bytes.HasPrefix(tr.Body, []byte("Something: interesting")) {
					t.Error("Wrong extra header line.")
				}
				return nil
			},
		},
		ForceTLS: true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestTwoExtraHeadersMakeMessageParsable(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		Hostname: "foobar.example.net",
		Handlers: []func(tr *Transaction) error{
			func(tr *Transaction) error {
				tr.AddHeader("Something1", "interesting 1")
				tr.AddHeader("Something2", "interesting 2")
				tr.AddReceivedLine()
				if !bytes.HasPrefix(tr.Body, []byte("Received: from localhost ([127.0.0.1]) by foobar.example.net with ESMTP;")) {
					t.Error("Wrong received line.")
				}
				msg, err := mail.ReadMessage(bytes.NewReader(tr.Body))
				if err != nil {
					t.Errorf("%s : while parsing email message", err)
					return err
				}
				if msg.Header.Get("Something1") != "interesting 1" {
					t.Errorf("Header Something is wrong: `%s` instead of `interesting 1`",
						msg.Header.Get("Something1"))
				}
				if msg.Header.Get("Something2") != "interesting 2" {
					t.Errorf("Header Something is wrong: `%s` instead of `interesting 1`",
						msg.Header.Get("Something1"))
				}
				return nil
			},
		},
		ForceTLS: true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}

	body := `
Date: Sun, 11 Jun 2023 19:49:29 +0300
To: scuba@vodolaz095.ru
From: scuba@vodolaz095.ru
Subject: test Sun, 11 Jun 2023 19:49:29 +0300
Message-Id: <20230611194929.017435@localhost>
X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/

This is a test mailing
`
	_, err = fmt.Fprintf(wc, body)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestHELO(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = cmd(c.Text, 502, "MAIL FROM:<test@example.org>"); err != nil {
		t.Errorf("MAIL before HELO didn't fail: %v", err)
	}
	if err = cmd(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = cmd(c.Text, 250, "MAIL FROM:<test@example.org>"); err != nil {
		t.Errorf("MAIL after HELO failed: %v", err)
	}
	if err = cmd(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("double HELO failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestLOGINAuth(t *testing.T) {
	addr, closer := runsslserver(t, &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysWorks,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = cmd(c.Text, 334, "AUTH LOGIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = cmd(c.Text, 502, "foo"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = cmd(c.Text, 334, "AUTH LOGIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = cmd(c.Text, 334, "Zm9v"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = cmd(c.Text, 502, "foo"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = cmd(c.Text, 334, "AUTH LOGIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = cmd(c.Text, 334, "Zm9v"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = cmd(c.Text, 235, "Zm9v"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestNullSender(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = cmd(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = cmd(c.Text, 250, "MAIL FROM:<>"); err != nil {
		t.Errorf("MAIL with null sender failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestNoBracketsSender(t *testing.T) {
	addr, closer := runserver(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = cmd(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = cmd(c.Text, 250, "MAIL FROM:test@example.org"); err != nil {
		t.Errorf("MAIL without brackets failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestErrors(t *testing.T) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Errorf("Cert load failed: %v", err)
	}
	server := &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysWorks,
	}
	addr, closer := runserver(t, server)
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = cmd(c.Text, 502, "AUTH PLAIN foobar"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = cmd(c.Text, 502, "AUTH PLAIN foobar"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Errorf("MAIL didn't fail")
	}
	if err = cmd(c.Text, 502, "STARTTLS"); err != nil {
		t.Errorf("STARTTLS didn't fail: %v", err)
	}
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = cmd(c.Text, 502, "AUTH UNKNOWN"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = cmd(c.Text, 502, "AUTH PLAIN foobar"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = cmd(c.Text, 502, "AUTH PLAIN Zm9vAGJhcg=="); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = cmd(c.Text, 334, "AUTH PLAIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = cmd(c.Text, 235, "Zm9vAGJhcgBxdXV4"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Errorf("Duplicate MAIL didn't fail")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestMalformedMAILFROM(t *testing.T) {
	sc := make([]CheckerFunc, 0)
	sc = append(sc, func(tr *Transaction, name string) error {
		if name != "test@example.org" {
			return ErrorSMTP{Code: 502, Message: "Denied"}
		}
		return nil
	})
	addr, closer := runserver(t, &Server{
		SenderCheckers: sc,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = cmd(c.Text, 250, "MAIL FROM: <test@example.org>"); err != nil {
		t.Errorf("MAIL FROM failed with extra whitespace: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestKarma(t *testing.T) {
	addr, closer := runserver(t, &Server{
		SenderCheckers: []CheckerFunc{
			func(transaction *Transaction, name string) error {
				if transaction.Karma() != 0 {
					t.Errorf("wrong initial karma")
				}
				if name == "scuba@vodolaz095.ru" {
					transaction.Love(1000)
				}
				return nil
			},
		},
		Handlers: []func(tr *Transaction) error{
			func(tr *Transaction) error {
				if tr.Karma() != 1000 {
					t.Errorf("not enough karma")
				}
				err := ErrorSMTP{
					Code:    555,
					Message: "karma",
				}
				if err.Error() != "555 karma" {
					t.Errorf("wrong error")
				}
				return err
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = c.Hello("mx.example.org")
	if err != nil {
		t.Errorf("sending helo command failed with %s", err)
	}
	err = c.Mail("scuba@vodolaz095.ru")
	if err != nil {
		t.Errorf("sending mail from command failed with %s", err)
	}
	err = c.Rcpt("example@example.org")
	if err != nil {
		t.Errorf("RCPT TO command command failed with %s", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() != "555 karma" {
			t.Errorf("wrong error returned")
		}
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("sending quit command failed with %s", err)
	}
}

func TestProxyNotEnabled(t *testing.T) {
	addr, closer := runserver(t, &Server{
		EnableProxyProtocol: false, // important
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}

	where := strings.Split(addr, ":")
	err = cmd(c.Text, 550, "PROXY TCP4 8.8.8.8 %s 443 %s", where[0], where[1])
	if err != nil {
		t.Errorf("sending proxy command enabled from the box - %s", err)
	}

	err = c.Hello("nobody.example.org")
	if err != nil {
		t.Errorf("sending helo command failed with %s", err)
	}

	err = c.Quit()
	if err != nil {
		t.Errorf("sending quit command failed with %s", err)
	}
}

func TestTLSListener(t *testing.T) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Errorf("Cert load failed: %v", err)
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	defer ln.Close()
	addr := ln.Addr().String()
	server := &Server{
		Authenticator: func(tr *Transaction, username, password string) error {
			if tr.TLS == nil {
				t.Error("didn't correctly set connection state on TLS connection")
			}
			return nil
		},
	}
	go func() {
		server.Serve(ln)
	}()
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Errorf("couldn't connect to tls socket: %v", err)
	}
	c, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		t.Errorf("couldn't create client: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = cmd(c.Text, 334, "AUTH PLAIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = cmd(c.Text, 235, "Zm9vAGJhcgBxdXV4"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestShutdown(t *testing.T) {
	t.Logf("Starting shutdown test")
	server := &Server{}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("Listen failed: %v", err)
	}
	srvres := make(chan error)
	go func() {
		t.Log("Starting server")
		srvres <- server.Serve(ln)
	}()
	// Connect a client
	c, err := smtp.Dial(ln.Addr().String())
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	// While the client connection is open, shut down the server (without
	// waiting for it to finish)
	err = server.Shutdown(false)
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
	// Verify that Shutdown() worked by attempting to connect another client
	_, err = smtp.Dial(ln.Addr().String())
	if err == nil {
		t.Errorf("Dial did not fail as expected")
	}
	if _, typok := err.(*net.OpError); !typok {
		t.Errorf("Dial did not return net.OpError as expected: %v (%T)", err, err)
	}
	// Wait for shutdown to complete
	shutres := make(chan error)
	go func() {
		t.Log("Waiting for server shutdown to finish")
		shutres <- server.Wait()
	}()
	// Slight delay to ensure Shutdown() blocks
	time.Sleep(250 * time.Millisecond)
	// Wait() should not have returned yet due to open client conn
	select {
	case shuterr := <-shutres:
		t.Errorf("Wait() returned early w/ error: %v", shuterr)
	default:
	}
	// Now close the client
	t.Log("Closing client connection")
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
	c.Close()

	// Wait for Wait() to return
	t.Log("Waiting for Wait() to return")
	select {
	case shuterr := <-shutres:
		if shuterr != nil {
			t.Errorf("Wait() returned error: %v", shuterr)
		}
	case <-time.After(15 * time.Second):
		t.Errorf("Timed out waiting for Wait() to return")
	}

	// Wait for Serve() to return
	t.Log("Waiting for Serve() to return")
	select {
	case srverr := <-srvres:
		if srverr != ErrServerClosed {
			t.Errorf("Serve() returned error: %v", srverr)
		}
	case <-time.After(15 * time.Second):
		t.Errorf("Timed out waiting for Serve() to return")
	}
}

func TestServeFailsIfShutdown(t *testing.T) {
	server := &Server{}
	err := server.Shutdown(true)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
	err = server.Serve(nil)
	if err != ErrServerClosed {
		t.Errorf("Serve() did not return ErrServerClosed: %v", err)
	}
}

func TestWaitFailsIfNotShutdown(t *testing.T) {
	server := &Server{}
	err := server.Wait()
	if err == nil {
		t.Errorf("Wait() did not fail as expected")
	}
}
