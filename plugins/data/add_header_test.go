package data

import (
	"bytes"
	"fmt"
	"net/smtp"
	"testing"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

func TestAddHeader(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			AddHeader("Something", "interesting"),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(tr *msmtpd.Transaction) error {
				t.Logf("Body:\n--------------\n%s\n-----------\n", string(tr.Body))

				if !bytes.Contains(tr.Body, []byte("Something: interesting")) {
					t.Errorf("extra header not present in body")
				}
				val, found := tr.Parsed.Header["Something"]
				if !found {
					t.Errorf("header not added to parsed body object")
				}
				if len(val) != 1 {
					t.Errorf("wrong number of headers added to parsed body object")
				}
				if val[0] != "interesting" {
					t.Errorf("wrong header added to parsed body object")
				}
				t.Logf(tr.Parsed.Header.Get("Something"))
				if tr.Parsed.Header.Get("Something") != "interesting" {
					t.Errorf("wrong header added to parsed body object")
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
	_, err = fmt.Fprint(wc, internal.MakeTestMessage("scuba@vodolaz095.ru", "scuba@vodolaz095.ru"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
}
