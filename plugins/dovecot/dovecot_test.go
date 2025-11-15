package dovecot

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

var username, password, rcptTo string

// tested on actual dovecot-2.3.16-3.el8.rpm from Centos 8 stream repos

func TestLoadFromEnvironment(t *testing.T) {
	username = os.Getenv("DOVECOT_USERNAME")
	if username == "" {
		t.Skipf("environment variable DOVECOT_USERNAME is not set")
	}
	password = os.Getenv("DOVECOT_PASSWORD")
	if password == "" {
		t.Skipf("environment variable DOVECOT_PASSWORD is not set")
	}
	rcptTo = os.Getenv("DOVECOT_RCPT_TO")
	if rcptTo == "" {
		t.Skipf("environment variable DOVECOT_RCPT_TO is not set")
	}
}

func TestDovecot_Exists(t *testing.T) {
	if rcptTo == "" {
		t.Skipf("skipping, because environment variable DOVECOT_RCPT_TO is not set")
	}

	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LmtpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	tr := msmtpd.Transaction{
		ID:        "dovecot_exists",
		StartedAt: time.Now(),
	}
	err := dvc.CheckRecipient(context.TODO(), &tr, &mail.Address{
		Name:    "who cares",
		Address: rcptTo,
	})
	if err != nil {
		t.Errorf("%s : while checking %s to exists", err, rcptTo)
	}
	err = dvc.CheckRecipient(context.TODO(), &tr, &mail.Address{
		Name:    "who cares",
		Address: "somebody@example.org",
	})
	if err != nil {
		if err.Error() != "521 i have no idea about recipient you want to deliver message to" {
			t.Errorf("%s : wrong error while checking non existent mailbox", err)
		}
	} else {
		t.Errorf("error is expected")
	}
}

func TestDovecot_Authenticate(t *testing.T) {
	if username == "" {
		t.Skipf("skipping, because environment variable DOVECOT_USERNAME is not set")
	}
	if password == "" {
		t.Skipf("skipping, because environment variable DOVECOT_PASSWORD is not set")
	}

	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LmtpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	tr := msmtpd.Transaction{
		ID:        "dovecot_authenticate",
		StartedAt: time.Now(),
	}
	err := dvc.Authenticate(context.TODO(), &tr, username, password)
	if err != nil {
		t.Errorf("%s : while checking %s to exists", err, rcptTo)
	}
	err = dvc.Authenticate(context.TODO(), &tr, "wrong_username", "wrong_password")
	if err != nil {
		if err.Error() != "521 authorization failed" {
			t.Errorf("%s : wrong error while checking wrong authentication", err)
		}
	} else {
		t.Errorf("error is expected")
	}
}

func TestDovecot_DeliverRcptTo(t *testing.T) {
	validMessage := internal.MakeTestMessage("somebody@example.org", rcptTo)
	if rcptTo == "" {
		t.Skipf("skipping, because environment variable DOVECOT_RCPT_TO is not set")
	}
	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LmtpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	tr := msmtpd.Transaction{
		ID:        "dovecot_deliver_rcpt",
		StartedAt: time.Now(),
		Body:      []byte(validMessage),
		MailFrom:  mail.Address{Name: "who cares", Address: rcptTo},
		RcptTo: []mail.Address{
			{Name: "who cares", Address: rcptTo},
			{Name: "somebody", Address: "somebody@example.org"},
		},
	}
	err := dvc.Deliver(context.TODO(), &tr)
	if err != nil {
		t.Errorf("%s : while delivering test message", err)
	}
}

func TestDovecot_DeliverAliases(t *testing.T) {
	validMessage := internal.MakeTestMessage("somebody@example.org", rcptTo)
	if rcptTo == "" {
		t.Skipf("skipping, because environment variable DOVECOT_RCPT_TO is not set")
	}
	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LmtpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	tr := msmtpd.Transaction{
		ID:        "dovecot_deliver_alias",
		StartedAt: time.Now(),
		Body:      []byte(validMessage),
		MailFrom:  mail.Address{Name: "who cares", Address: rcptTo},
		Aliases: []mail.Address{
			{Name: "who cares", Address: rcptTo},
			{Name: "somebody", Address: "somebody@example.org"},
		},
	}
	err := dvc.Deliver(context.TODO(), &tr)
	if err != nil {
		t.Errorf("%s : while delivering test message", err)
	}
}

func TestDovecotIntegration(t *testing.T) {
	if username == "" {
		t.Skipf("skipping, because environment variable DOVECOT_USERNAME is not set")
	}
	if password == "" {
		t.Skipf("skipping, because environment variable DOVECOT_PASSWORD is not set")
	}
	if rcptTo == "" {
		t.Skipf("skipping, because environment variable DOVECOT_RCPT_TO is not set")
	}
	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LmtpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	validMessage := internal.MakeTestMessage("sender@example.org", rcptTo)
	addr, closer := msmtpd.RunTestServerWithTLS(t, &msmtpd.Server{
		Hostname:      "localhost", // required for authentication
		ForceTLS:      true,
		Authenticator: dvc.Authenticate,
		RecipientCheckers: []msmtpd.RecipientChecker{
			dvc.CheckRecipient,
		},
		DataHandlers: []msmtpd.DataHandler{
			dvc.Deliver,
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
		return
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
		return
	}
	t.Logf("HELO PASSED")
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
		return
	}
	t.Logf("STARTTLS PASSED")
	err = c.Auth(smtp.PlainAuth("", username, password, "127.0.0.1"))
	if err != nil {
		t.Errorf("%s : while performing authentication", err)
		return
	}
	t.Logf("AUTH PASSED")
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
		return
	}
	t.Logf("MAIL FROM PASSED")
	if err = c.Rcpt(rcptTo); err != nil {
		t.Errorf("Rcpt failed: %v", err)
		return
	}
	t.Logf("RCPT TO PASSED")

	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
		return
	}
	_, err = fmt.Fprint(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
		return
	}
	t.Logf("DATA PASSED")

	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
}
