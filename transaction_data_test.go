package msmtpd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"testing"

	"msmtpd/internal"
)

func TestMaxMessageSize(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
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

func TestEnvelopeReceived(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunTestServerWithTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
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
	addr, closer := RunTestServerWithTLS(t, &Server{
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

	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
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

func TestMalformedMessageBody(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
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

func TestInterruptedDATA(t *testing.T) {
	handlers := make([]DataHandler, 0)
	handlers = append(handlers, func(tr *Transaction) error {
		t.Error("Accepted DATA despite disconnection")
		return nil
	})
	addr, closer := RunTestServerWithoutTLS(t, &Server{
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
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	c.Close()
}
