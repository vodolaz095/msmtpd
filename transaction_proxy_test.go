package msmtpd

import (
	"net/smtp"
	"strings"
	"testing"

	"msmtpd/internal"
)

func TestProxyNotEnabled(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		EnableProxyProtocol: false, // important
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}

	where := strings.Split(addr, ":")
	err = internal.DoCommand(c.Text, 550, "PROXY TCP4 8.8.8.8 %s 443 %s", where[0], where[1])
	if err != nil {
		t.Errorf("sending proxy command enabled from the box - %s", err)
	}

	err = c.Hello("nobody.example.org")
	if err != nil {
		t.Errorf("sending helo command failed with %s", err)
	}

	err = c.Quit()
	if err != nil {
		t.Errorf("sending quit command failed with %s", err)
	}
}
