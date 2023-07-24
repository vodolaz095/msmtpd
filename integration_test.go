package msmtpd

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSMTP(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	addr, closer := RunServerWithoutTLS(t, server)
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
	addr, closer := RunServerWithTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunServerWithTLS(t, &Server{
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
	addr, closer := RunServerWithTLS(t, &Server{
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
	addr, closer := RunServerWithTLS(t, &Server{
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
	cc := make([]ConnectionChecker, 0)
	cc = append(cc, func(tr *Transaction) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
		ConnectionCheckers: cc,
	})
	defer closer()
	if _, err := smtp.Dial(addr); err == nil {
		t.Error("Dial succeeded despite ConnectionCheck")
	}
}

func TestConnectionCheckSimpleError(t *testing.T) {
	cc := make([]ConnectionChecker, 0)
	cc = append(cc, func(tr *Transaction) error {
		return errors.New("Denied")
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
		ConnectionCheckers: cc,
	})
	defer closer()
	if _, err := smtp.Dial(addr); err == nil {
		t.Error("Dial succeeded despite ConnectionCheck")
	}
}

func TestHELOCheck(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{
		HeloCheckers: []HelloChecker{
			func(transaction *Transaction) error {
				name := transaction.HeloName
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
	sc := make([]SenderChecker, 0)
	sc = append(sc, func(tr *Transaction) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	rc := make([]RecipientChecker, 0)
	rc = append(rc, func(tr *Transaction, name *mail.Address) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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

func TestDataHandler(t *testing.T) {
	handlers := make([]DataHandler, 0)
	handlers = append(handlers, func(tr *Transaction) error {
		if len(tr.PTRs) != 1 {
			t.Errorf("wrong length of PTR records for localhost - %v", len(tr.PTRs))
		}
		if tr.PTRs[0] != "localhost" {
			t.Errorf("wrong PTR record for localhost - %v", tr.PTRs)
		}
		if tr.MailFrom.Address != "sender@example.org" {
			t.Errorf("Unknown sender: %v", tr.MailFrom)
		}
		if len(tr.RcptTo) != 1 {
			t.Errorf("Too many recipients: %d", len(tr.RcptTo))
		}
		if tr.RcptTo[0].Address != "recipient@example.net" {
			t.Errorf("Unknown recipient: %v", tr.RcptTo[0].Address)
		}
		if !strings.Contains(string(tr.Body), "This is test message send from sender@example.org to recipient@example.net on") {
			t.Errorf("Wrong message body: %v", string(tr.Body))
		}
		return nil
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
		DataHandlers: handlers,
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	handlers := make([]DataHandler, 0)
	handlers = append(handlers, func(tr *Transaction) error {
		return ErrorSMTP{Code: 550, Message: "Rejected"}
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
		DataHandlers: handlers,
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err == nil {
		t.Error("Unexpected RCPT success")
	}
}

func TestDATAbeforeMailFrom(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if _, err = c.Data(); err == nil {
		t.Error("Data accepted despite no sender")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestDATAbeforeRCPT(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	handlers := make([]DataHandler, 0)
	handlers = append(handlers, func(tr *Transaction) error {
		t.Error("Accepted DATA despite disconnection")
		return nil
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
		DataHandlers: handlers,
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	c.Close()
}

func TestContext(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	addr, closer := RunServerWithoutTLS(t, &Server{
		ConnectionCheckers: []ConnectionChecker{
			func(transaction *Transaction) error {
				ctx := transaction.Context()
				t.Logf("context is extracted!")
				go func() {
					t.Logf("starting background goroutine being terminated with context")
					<-ctx.Done()
					wg.Done()
					t.Logf("context is terminated")
				}()
				return nil
			},
		},
	})

	defer closer()
	cm, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	t.Logf("closing connection, so context should be closed too...")
	err = cm.Close()
	if err != nil {
		t.Error(err)
	}
	t.Logf("waiting for context to be closed...")
	wg.Wait()
	t.Logf("context is closed")
}

func TestMeta(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{
		MaxConnections: 1,
		HeloCheckers: []HelloChecker{
			func(transaction *Transaction) error {
				name := transaction.HeloName
				transaction.SetFact("something", name)
				transaction.Incr("int64", 1)
				transaction.Incr("float64", 1.1)
				transaction.LogWarn("something")
				transaction.SetFlag("heloCheckerFired")
				return nil
			},
		},
		SenderCheckers: []SenderChecker{
			func(transaction *Transaction) error {
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
				if !transaction.IsFlagSet("heloCheckerFired") {
					t.Errorf("flag heloCheckerFired is not set")
				}
				transaction.UnsetFlag("heloCheckerFired")
				return nil
			},
		},
		RecipientCheckers: []RecipientChecker{
			func(transaction *Transaction, _ *mail.Address) error {
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
				if transaction.IsFlagSet("heloCheckerFired") {
					t.Errorf("flag heloCheckerFired is set")
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
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	addr, closer := RunServerWithTLS(t, &Server{
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	sc := make([]SenderChecker, 0)
	sc = append(sc, func(tr *Transaction) error {
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
	addr, closer := RunServerWithoutTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunServerWithTLS(t, &Server{
		Hostname: "foobar.example.net",
		DataHandlers: []DataHandler{
			func(tr *Transaction) error {
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunServerWithTLS(t, &Server{
		Hostname: "foobar.example.net",
		DataHandlers: []DataHandler{
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunServerWithTLS(t, &Server{
		Hostname: "foobar.example.net",
		DataHandlers: []DataHandler{
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

	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	addr, closer := RunServerWithTLS(t, &Server{
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	addr, closer := RunServerWithoutTLS(t, server)
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
	sc := make([]SenderChecker, 0)
	sc = append(sc, func(tr *Transaction) error {
		if tr.MailFrom.Address != "test@example.org" {
			return ErrorSMTP{Code: 502, Message: "Denied"}
		}
		return nil
	})
	addr, closer := RunServerWithoutTLS(t, &Server{
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

func TestUnparsableMessageBody(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{})
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
	_, err = fmt.Fprintf(wc, "this is nonsense")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() != "521 Stop sending me this nonsense, please!" {
			t.Errorf("%s : while closing message body", err)
		}
	} else {
		t.Errorf("error not returned for sending malformed message body")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
}

func TestKarma(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{
		SenderCheckers: []SenderChecker{
			func(transaction *Transaction) error {
				karma := transaction.Karma()
				if karma != commandExecutedProperly { // because HELO passed
					t.Errorf("wrong initial karma %v", karma)
				}
				if transaction.MailFrom.Address == "scuba@vodolaz095.ru" {
					transaction.Love(1000)
				}
				return nil
			},
		},
		DataHandlers: []DataHandler{
			func(tr *Transaction) error {
				if tr.Karma() < 1000 {
					t.Errorf("not enough karma. Required at least 1000. Actual: %v", tr.Karma())
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
	_, err = fmt.Fprintf(wc, MakeTestMessage("sender@example.org", "recipient@example.net"))
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

func TestCloseHandlers(t *testing.T) {
	wg := sync.WaitGroup{}
	var closeHandler1Called bool
	var closeHandler2Called bool
	addr, closer := RunServerWithoutTLS(t, &Server{
		ConnectionCheckers: []ConnectionChecker{
			func(transaction *Transaction) error {
				t.Logf("Giving 2 wg to transaction %s", transaction.ID)
				wg.Add(2)
				return nil
			},
		},
		CloseHandlers: []CloseHandler{
			func(transaction *Transaction) error {
				t.Logf("Closing transaction %s by 1st handler", transaction.ID)
				closeHandler1Called = true
				wg.Done()
				return nil
			},
			func(transaction *Transaction) error {
				t.Logf("Closing transaction %s by 2nd handler", transaction.ID)
				closeHandler2Called = true
				wg.Done()
				return fmt.Errorf("who cares")
			},
		},
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = c.Close()
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	wg.Wait()
	if !closeHandler1Called {
		t.Errorf("close handler 1 is not called")
	}
	if !closeHandler2Called {
		t.Errorf("close handler 2 is not called")
	}
}

func TestProxyNotEnabled(t *testing.T) {
	addr, closer := RunServerWithoutTLS(t, &Server{
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
