package msmtpd

import (
	"fmt"
	"net/smtp"
	"testing"

	"msmtpd/internal"
)

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
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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
	err = internal.DoCommand(c.Text, 220, "XCLIENT NAME=ignored ADDR=42.42.42.42 PORT=4242 PROTO=SMTP HELO=new.example.net LOGIN=newusername")
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
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
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
