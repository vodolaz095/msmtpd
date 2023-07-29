package data

import (
	"fmt"
	"net/smtp"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestBodyParseAndCheckHeadersOK(t *testing.T) {
	validMessage := `Date: Sun, 11 Jun 2023 19:49:29 +0300
To: scuba@vodolaz095.ru
From: scuba@vodolaz095.ru
Subject: test Sun, 11 Jun 2023 19:49:29 +0300
Message-Id: <20230611194929.017435@localhost>
X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/

This is a test mailing
`

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			CheckHeaders(DefaultHeadersToRequire),
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
	if err = c.Mail("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
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

func TestBodyParseAndCheckHeadersMalformed(t *testing.T) {
	malformedMessage := `This is a test mailing`
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			CheckHeaders(DefaultHeadersToRequire),
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
	if err = c.Mail("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, malformedMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() == "521 Stop sending me this nonsense, please!" {
			t.Logf("proper error is thrown")
			return
		} else {
			t.Errorf("Data close failed with wrong error %v", err)
		}
	}
	t.Errorf("error not thrown")
}

func TestBodyParseAndCheckHeadersMissingHeaders(t *testing.T) {
	messageWithoutFrom := `Date: Sun, 11 Jun 2023 19:49:29 +0300
To: scuba@vodolaz095.ru
Subject: test Sun, 11 Jun 2023 19:49:29 +0300
Message-Id: <20230611194929.017435@localhost>
X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/

This is a test mailing
`

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			CheckHeaders(DefaultHeadersToRequire),
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
	if err = c.Mail("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, messageWithoutFrom)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() == "521 I cannot parse your message. Do not send me this particular message in future, please, i will never accept it. Thanks in advance!" {
			t.Logf("proper error is thrown")
			return
		} else {
			t.Errorf("Data close failed with wrong error %v", err)
		}
	}
	t.Errorf("error not thrown")
}
