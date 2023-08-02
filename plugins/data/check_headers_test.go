package data

import (
	"bytes"
	"fmt"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
)

func TestBodyParseAndCheckHeadersOK(t *testing.T) {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Date: %s\n", time.Now().Format(time.RFC1123Z))
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	buf.WriteString("From: scuba@vodolaz095.ru\n")
	buf.WriteString("Subject: test with all headers ok\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing")

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
	_, err = fmt.Fprintf(wc, buf.String())
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

func TestBodyParseAndCheckHeadersMissingMandatoryHeaderFrom(t *testing.T) {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Date: %s\n", time.Now().Format(time.RFC1123Z))
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	// buf.WriteString("From: scuba@vodolaz095.ru\n") // IMPORTANT
	buf.WriteString("Subject: from not present\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing without FROM header")

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
	_, err = fmt.Fprintf(wc, buf.String())
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

func TestBodyParseAndCheckHeadersMissingMandatoryHeaderDate(t *testing.T) {
	buf := bytes.NewBufferString("")
	//	fmt.Fprintf(buf, "Date: %s\n", time.Now().Format(time.RFC1123Z)) // IMPORTANT
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	buf.WriteString("From: scuba@vodolaz095.ru\n")
	buf.WriteString("Subject: date is missing\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing without DATE header")

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
	_, err = fmt.Fprintf(wc, buf.String())
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

func TestBodyParseAndCheckHeadersMissingRequiredSubject(t *testing.T) {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Date: %s\n", time.Now().Format(time.RFC1123Z))
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	buf.WriteString("From: scuba@vodolaz095.ru\n")
	// buf.WriteString("Subject: subject is not mandatory, but it is missing\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing without SUBJECT header")

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
	_, err = fmt.Fprintf(wc, buf.String())
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

func TestBodyParseAndCheckHeadersDuplicate(t *testing.T) {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Date: %s\n", time.Now().Format(time.RFC1123Z))
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	buf.WriteString("From: scuba@vodolaz095.ru\n")
	buf.WriteString("Subject: test with duplicate headers\n")
	buf.WriteString("Subject: duplicate subject\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing with duplicate subject")

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
	_, err = fmt.Fprintf(wc, buf.String())
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

func TestBodyParseAndCheckHeadersDateTooOld(t *testing.T) {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Date: %s\n", time.Now().AddDate(-1, 1, 1).Format(time.RFC1123Z))
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	buf.WriteString("From: scuba@vodolaz095.ru\n")
	buf.WriteString("Subject: test date too old\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing with date too old")

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
	_, err = fmt.Fprintf(wc, buf.String())
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

func TestBodyParseAndCheckHeadersDateToFarInFuture(t *testing.T) {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Date: %s\n", time.Now().AddDate(0, 1, 1).Format(time.RFC1123Z))
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	buf.WriteString("From: scuba@vodolaz095.ru\n")
	buf.WriteString("Subject: test date in future\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing with date in future")

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
	_, err = fmt.Fprintf(wc, buf.String())
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

func TestBodyParseAndCheckHeadersDateMalformed(t *testing.T) {
	buf := bytes.NewBufferString("")
	buf.WriteString("Date: сегодня, после обеда\n") // yes
	buf.WriteString("To: scuba@vodolaz095.ru\n")
	buf.WriteString("From: scuba@vodolaz095.ru\n")
	buf.WriteString("Subject: test with strange date\n")
	buf.WriteString("Message-Id: <20230611194929.017435@localhost>\n")
	buf.WriteString("\n\nThis is a test mailing with strange date")

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
	_, err = fmt.Fprintf(wc, buf.String())
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
