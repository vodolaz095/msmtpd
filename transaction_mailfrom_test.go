package msmtpd

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"testing"

	"github.com/vodolaz095/msmtpd/internal"
)

func TestLongLineInMailFrom(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
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

func TestNoBracketsSender(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "MAIL FROM:test@example.org"); err != nil {
		t.Errorf("MAIL without brackets failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestInvalidSender(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("invalid@@example.org"); err == nil {
		t.Error("Unexpected MAIL success")
	}
	err = c.Close()
	if err != nil {
		t.Errorf("%s : while closing transaction", err)
	}
}

func TestNullSender(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "MAIL FROM:<>"); err != nil {
		t.Errorf("MAIL with null sender failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestSenderCheck(t *testing.T) {
	sc := make([]SenderChecker, 0)
	sc = append(sc, func(_ context.Context, tr *Transaction) error {
		if tr.MailFrom.Address != "sender@example.org" {
			t.Errorf("wrong sender %s", tr.MailFrom.String())
		}
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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
	if err = c.Quit(); err != nil {
		t.Errorf("%s : while quiting", err)
	}
}

func TestMailFromWithExtraSpace(t *testing.T) {
	sc := make([]SenderChecker, 0)
	sc = append(sc, func(_ context.Context, tr *Transaction) error {
		if tr.MailFrom.Address != "test@example.org" {
			return ErrorSMTP{Code: 502, Message: "Denied"}
		}
		return nil
	})
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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
	err = internal.DoCommand(c.Text, 502, "MAIL")
	if err != nil {
		t.Errorf("%s : while sending malformed command", err)
	}
	if err = internal.DoCommand(c.Text, 250, "MAIL FROM: <test@example.org>"); err != nil {
		t.Errorf("MAIL FROM failed with extra whitespace: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestRcptToBeforeMAIL(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err == nil {
		t.Error("Unexpected RCPT success")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("%s : while quiting", err)
	}
}

func TestDataBeforeMailFrom(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
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
