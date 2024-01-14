package deliver

import (
	"fmt"
	"net/smtp"
	"os"
	"testing"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

var testLMTPRecipient string

func TestViaLocalMailTransferProtocol(t *testing.T) {
	testLMTPRecipient = os.Getenv("TEST_LMTP_RECIPIENT")
	if testLMTPRecipient == "" {
		t.Skipf("TEST_LMTP_RECIPIENT is not set")
	}
	validMessage := internal.MakeTestMessage("something@localhost", testLMTPRecipient)
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataHandlers: []msmtpd.DataHandler{
			ViaLocalMailTransferProtocol(LMTPOptions{
				Network: "unix",
				Address: "/var/run/dovecot/lmtp",
				LHLO:    "localhost",
				Timeout: DefaultTimeout,
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
	if err = c.Mail("something@localhost"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt(testLMTPRecipient); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprint(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
}
