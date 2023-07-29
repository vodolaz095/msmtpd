package msmtpd

import (
	"crypto/tls"
	"net/mail"
	"net/smtp"
	"testing"

	"github.com/vodolaz095/msmtpd/internal"
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
	err = c.Close()
	if err != nil {
		t.Errorf("%s : while closing", err)
	}
}

func TestMalformedRcpt(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	err = internal.DoCommand(c.Text, 502, "RCPT")
	if err != nil {
		t.Errorf("%s : while making malformed rcpt to", err)
	}
	err = internal.DoCommand(c.Text, 502, "RCPT FOR")
	if err != nil {
		t.Errorf("%s : while making malformed rcpt to", err)
	}
	err = c.Close()
	if err != nil {
		t.Errorf("%s : while closing", err)
	}
}

func TestRCPTinWrongOrder(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
		ForceTLS:      true,
		Authenticator: AuthenticatorForTestsThatAlwaysWorks,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = c.Rcpt("bill.gates@microsoft.com")
	if err != nil {
		if err.Error() != "502 Please introduce yourself first." {
			t.Errorf("%s : wrong error while helo not called", err)
		}
	} else {
		t.Error("error not thrown when DATA called before HELO")
	}
	err = c.Hello("localhost")
	if err != nil {
		t.Errorf("%s : while sending HELO", err)
	}
	err = c.Rcpt("bill.gates@microsoft.com")
	if err != nil {
		if err.Error() != "502 Please turn on TLS by issuing a STARTTLS command." {
			t.Errorf("%s : wrong error while STARTTLS not called", err)
		}
	} else {
		t.Error("error not thrown when DATA called before STARTTLS")
	}
	err = c.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Errorf("%s : while sending STARTTLS", err)
	}
	err = c.Rcpt("bill.gates@microsoft.com")
	if err != nil {
		if err.Error() != "530 Authentication Required." {
			t.Errorf("%s : wrong error while STARTTLS not called", err)
		}
	} else {
		t.Error("error not thrown when DATA called before AUTH")
	}
	err = c.Auth(smtp.PlainAuth("", "who", "cares", "127.0.0.1"))
	if err != nil {
		t.Errorf("%s : while sending AUTH", err)
	}
	err = c.Rcpt("bill.gates@microsoft.com")
	if err != nil {
		if err.Error() != "502 It seems you haven't called MAIL FROM in order to explain who sends your message." {
			t.Errorf("%s : wrong error while MAIL FROM not called", err)
		}
	} else {
		t.Error("error not thrown when DATA called before MAIL FROM")
	}
	err = c.Mail("somebody@example.org")
	if err != nil {
		t.Errorf("%s : while sending MAILFROM", err)
	}
	err = c.Rcpt("bill.gates@microsoft.com")
	if err != nil {
		t.Errorf("%s : while sending RCPT TO", err)
	}
	err = c.Close()
	if err != nil {
		t.Errorf("%s : while closing", err)
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
