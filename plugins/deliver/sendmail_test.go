package deliver

import (
	"fmt"
	"net/smtp"
	"os"
	"testing"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

var sendMailRecipient string

func TestViaSendmail(t *testing.T) {
	sendMailRecipient = os.Getenv("TEST_SENDMAIL_RECIPIENT")
	if sendMailRecipient == "" {
		t.Skipf("TEST_SENDMAIL_RECIPIENT is not set")
	}

	validMessage := internal.MakeTestMessage("something@localhost", sendMailRecipient)
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataHandlers: []msmtpd.DataHandler{
			ViaSendmail(&SendmailOptions{
				PathToExecutable: "", // let us guess
				UseMinusTFlag:    false,
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
	if err = c.Rcpt(sendMailRecipient); err != nil {
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
