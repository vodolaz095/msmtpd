package msmtpd

import (
	"bytes"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// makeSmugglingMessage makes test email message with attempt to smuggle
func makeSmugglingMessage(separator, from string, to ...string) string {
	now := time.Now()
	buh := bytes.NewBufferString("Date: " + now.Format(time.RFC1123Z) + "\r\n")
	buh.WriteString("From: " + from + "\r\n")
	buh.WriteString("To: " + strings.Join(to, ",") + "\r\n")
	buh.WriteString(fmt.Sprintf("Subject: Test email send on %s\r\n", now.Format(time.RFC1123Z)))
	buh.WriteString(fmt.Sprintf("Message-Id: <%s@localhost>\r\n", now.Format("20060102150405")))
	buh.WriteString("\r\n")
	buh.WriteString(fmt.Sprintf("This is test message send from %s to %s on %s\r\n",
		from, strings.Join(to, ","), now.Format(time.Stamp),
	))
	buh.WriteString(separator)
	buh.WriteString("HELO lol\r\n")
	buh.WriteString("MAIL FROM:<bad@example.org>\r\n")
	buh.WriteString("RCPT TO:<bad@example.org>\r\n")
	return buh.String()
}

func TestSMTPSmugglingNotWorks(t *testing.T) {
	type testCase struct {
		separator string
	}

	// list of separators
	// https://github.com/The-Login/SMTP-Smuggling-Tools/blob/235cbf27ec66437f767013ae9f37c56a30648932/smtp_smuggling_scanner.py#L13
	testCases := []testCase{
		{"\r\n.\r\n"}, // correct one
		{"\n.\n"},
		{"\n.\r"},
		{"\r.\n"},
		{"\r.\r"},
		{"\n.\r\n"},
		{"\r.\r\n"},
		{"\r\n\x00.\r\n"},
		{"\r\n.\r\r\n"},
		{"\r\r\n.\r\r\n"},
		{"\r\n\x00.\r\n"},
	}

	for i := range testCases {
		t.Run(fmt.Sprintf("Case %v with separator %q", i, testCases[i].separator), func(tt *testing.T) {
			var numberOfMessagesAccepted uint32
			addr, closer := RunTestServerWithoutTLS(tt, &Server{
				HeloCheckers: []HelloChecker{
					func(tr *Transaction) error {
						if tr.HeloName == "lol" {
							tt.Errorf("smuggling encountered, helo accepted from message body")
						}
						return nil
					},
				},
				SenderCheckers: []SenderChecker{
					func(tr *Transaction) error {
						if tr.MailFrom.Address == "bad@example.org" {
							tt.Errorf("smuggling encountered, MAIL FROM accepted from message body")
						}
						return nil
					},
				},
				RecipientCheckers: []RecipientChecker{
					func(tr *Transaction, recipient *mail.Address) error {
						if recipient.Address == "bad@example.org" {
							tt.Errorf("smuggling encountered, RCPT TO accepted from message body")
						}
						return nil
					},
				},
				DataHandlers: []DataHandler{
					func(tr *Transaction) error {
						atomic.AddUint32(&numberOfMessagesAccepted, 1)
						tt.Log("Message content: ", string(tr.Body))
						return nil
					},
				},
			})
			defer closer()
			c, err := smtp.Dial(addr)
			if err != nil {
				tt.Errorf("Dial failed: %v", err)
				return
			}
			err = c.Hello("localhost")
			if err != nil {
				tt.Errorf("helo failed: %v", err)
				return
			}
			if err = c.Mail("sender@example.org"); err != nil {
				tt.Errorf("MAIL failed: %v", err)
				return
			}
			if err = c.Rcpt("recipient@example.net"); err != nil {
				tt.Errorf("RCPT failed: %v", err)
				return
			}
			wc, err := c.Data()
			if err != nil {
				tt.Errorf("error calling data: %v", err)
				return
			}
			n, err := fmt.Fprint(c.Text.W, makeSmugglingMessage(
				testCases[i].separator,
				"sender@example.org",
				"recipient@example.org",
			))
			if err != nil {
				tt.Errorf("error writing data: %v", err)
				return
			}
			tt.Logf("%v bytes written", n)
			err = wc.Close()
			if err != nil {
				tt.Errorf("error closing channel: %v", err)
				return
			}
			if numberOfMessagesAccepted != 1 {
				tt.Errorf("smuggling encountered after connection close")
			}
			tt.Logf("Case %v finished for separator %q: %v messages accepted", i, testCases[i].separator,
				numberOfMessagesAccepted)
		})
	}
}
