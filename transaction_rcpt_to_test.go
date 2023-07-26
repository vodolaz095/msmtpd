package msmtpd

import (
	"net/mail"
	"net/smtp"
	"testing"
)

func TestRecipientCheck(t *testing.T) {
	rc := make([]RecipientChecker, 0)
	rc = append(rc, func(tr *Transaction, name *mail.Address) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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

func TestMaxRecipients(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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

func TestInvalidRecipient(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
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

func TestDataCalledBeforeRCPT(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
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
