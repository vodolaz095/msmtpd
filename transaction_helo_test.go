package msmtpd

import (
	"net/smtp"
	"testing"

	"msmtpd/internal"
)

func TestHELOCheck(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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

func TestHELO(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "MAIL FROM:<test@example.org>"); err != nil {
		t.Errorf("MAIL before HELO didn't fail: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "MAIL FROM:<test@example.org>"); err != nil {
		t.Errorf("MAIL after HELO failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 250, "HELO localhost"); err != nil {
		t.Errorf("double HELO failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestInvalidHelo(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello(""); err == nil {
		t.Error("Unexpected HELO success")
	}
}
