package rspamd

import (
	"fmt"
	"net/smtp"
	"os"
	"testing"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

var testRspamdURL, testRspamdPassword string

func TestRspamdEnv(t *testing.T) {
	if os.Getenv("TEST_RSPAMD_URL") == "" {
		t.Errorf("Environment variable TEST_RSPAMD_URL is not set")
	} else {
		testRspamdURL = os.Getenv("TEST_RSPAMD_URL")
	}
	if os.Getenv("TEST_RSPAMD_PASSWORD") == "" {
		t.Errorf("Environment variable TEST_RSPAMD_PASSWORD is not set")
	} else {
		testRspamdPassword = os.Getenv("TEST_RSPAMD_PASSWORD")
	}
}

func TestCheckPyRSPAMD(t *testing.T) {
	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      testRspamdURL,
				Password: testRspamdPassword,
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
				return nil
			},
		},
	})
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
}