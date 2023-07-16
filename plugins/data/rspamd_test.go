package data

import (
	"fmt"
	"net/smtp"
	"os"
	"testing"

	"msmtpd"
)

var testRspamdUrl, testRspamdPassword string

func TestEnv(t *testing.T) {
	if os.Getenv("TEST_RSPAMD_URL") == "" {
		t.Errorf("Environment variable TEST_RSPAMD_URL is not set")
	} else {
		testRspamdUrl = os.Getenv("TEST_RSPAMD_URL")
	}
	if os.Getenv("TEST_RSPAMD_PASSWORD") == "" {
		t.Errorf("Environment variable TEST_RSPAMD_PASSWORD is not set")
	} else {
		testRspamdPassword = os.Getenv("TEST_RSPAMD_PASSWORD")
	}
}

func TestCheckPyRSPAMD(t *testing.T) {
	validMessage := `Date: Sun, 11 Jun 2023 19:49:29 +0300
To: scuba@vodolaz095.ru
From: scuba@vodolaz095.ru
Subject: test Sun, 11 Jun 2023 19:49:29 +0300
Message-Id: <20230611194929.017435@localhost>
X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/

This is a test mailing
`
	addr, closer := runserver(t, &msmtpd.Server{
		Logger: &testLogger{},
		DataHandlers: []msmtpd.DataHandler{
			CheckPyRSPAMD(RspamdOpts{
				Url:      testRspamdUrl,
				Password: testRspamdPassword,
			}),
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
