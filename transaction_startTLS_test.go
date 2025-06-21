package msmtpd

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"testing"

	"github.com/vodolaz095/msmtpd/internal"
)

func TestSTARTTLS(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
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
		t.Error("Mail worked before TLS with ForceTLS")
	}
	if err = internal.DoCommand(c.Text, 220, "STARTTLS"); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "foobar"); err == nil {
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
	_, err = fmt.Fprint(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
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
