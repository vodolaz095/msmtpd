package deliver

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"testing"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

var testProxyServer, testProxyUsername, testProxyPassword, testProxySender, testProxyRecipient string

func TestDeliverViaSMTPProxy(t *testing.T) {
	testProxyUsername = os.Getenv("TEST_PROXY_USERNAME")
	testProxyServer = os.Getenv("TEST_PROXY_SERVER")
	testProxyPassword = os.Getenv("TEST_PROXY_PASSWORD")
	testProxySender = os.Getenv("TEST_PROXY_SENDER")
	testProxyRecipient = os.Getenv("TEST_PROXY_RECIPIENT")

	if testProxyServer == "" {
		t.Skipf("environment variable TEST_PROXY_SERVER is not set")
	}

	validMessage := internal.MakeTestMessage(testProxySender, testProxyRecipient)
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataHandlers: []msmtpd.DataHandler{
			ViaSMTPProxy(SMTPProxyOptions{
				Network: "tcp",
				Address: testProxyServer + ":587",
				HELO:    "localhost",
				TLS: &tls.Config{
					ServerName: testProxyServer,
				},
				Auth:     smtp.PlainAuth("", testProxyUsername, testProxyPassword, "localhost"),
				MailFrom: "",
				RcptTo:   nil,
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
	if err = c.Mail(testProxySender); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt(testProxyRecipient); err != nil {
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
