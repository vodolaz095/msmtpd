package sender

import (
	"fmt"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
)

func TestAcceptMailFromDomains(t *testing.T) {
	cases := make(map[string]error, 0)

	wlErr := fmt.Errorf(`521 "I'm sorry, but your email address is not in whitelist"`)

	cases["thisIsNotAEmail"] = fmt.Errorf(`502 "Malformed e-mail address"`)
	cases["a@example.org"] = nil
	cases["a@vodolaz095.ru"] = nil
	cases["a@gmail.com"] = wlErr
	cases["b@gmail.com"] = wlErr

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		SenderCheckers: []msmtpd.SenderChecker{
			AcceptMailFromDomains([]string{ // it should have higher priority
				"example.org",
				"vodolaz095.ru",
			}),
		},
	})
	for k, v := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		if err = c.Hello("localhost"); err != nil {
			t.Errorf("HELO failed: %v", err)
		}
		err = c.Mail(k)
		if err != nil {
			if v != nil {
				if err.Error() != v.Error() {
					t.Errorf("wrong error checking %s. Expected: %s. Received: %s", k, v, err)
				}
				continue
			}
			t.Errorf("unexpected error checking %s - %s", k, err)
		}
		err = c.Quit()
		if err != nil {
			t.Errorf("%s : while closing connection", err)
		}
	}
	time.Sleep(time.Second)
	closer()
}

func TestAcceptMailFromAddresses(t *testing.T) {
	cases := make(map[string]error, 0)
	wlErr := fmt.Errorf(`521 "I'm sorry, but your email address is not in whitelist"`)

	cases["thisIsNotAEmail"] = fmt.Errorf(`502 "Malformed e-mail address"`)
	cases["a@gmail.com"] = nil
	cases["b@gmail.com"] = nil
	cases["d@gmail.com"] = wlErr
	cases["e@gmail.com"] = wlErr

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		SenderCheckers: []msmtpd.SenderChecker{
			AcceptMailFromAddresses([]string{ // it should have higher priority
				"a@gmail.com",
				"b@gmail.com",
			}),
		},
	})
	for k, v := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		if err = c.Hello("localhost"); err != nil {
			t.Errorf("HELO failed: %v", err)
		}
		err = c.Mail(k)
		if err != nil {
			if v != nil {
				if err.Error() != v.Error() {
					t.Errorf("wrong error checking %s. Expected: %s. Received: %s", k, v, err)
				}
				continue
			}
			t.Errorf("unexpected error checking %s - %s", k, err)
			continue
		}
		err = c.Quit()
		if err != nil {
			t.Errorf("%s : while closing connection", err)
		}
	}
	time.Sleep(time.Second)
	closer()
}

func TestAcceptMailFromDomainsOrAddresses(t *testing.T) {
	cases := make(map[string]error, 0)

	wlErr := fmt.Errorf(`521 "I'm sorry, but your email address is not in whitelist"`)
	cases["thisIsNotAEmail"] = fmt.Errorf("502 Malformed e-mail address")
	cases["a@example.org"] = nil
	cases["a@vodolaz095.ru"] = nil
	cases["a@gmail.com"] = wlErr
	cases["b@gmail.com"] = wlErr

	cases["a@gmail.com"] = nil
	cases["b@gmail.com"] = nil
	cases["d@gmail.com"] = wlErr
	cases["e@gmail.com"] = wlErr
	cases["info@yandex.ru"] = wlErr

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		SenderCheckers: []msmtpd.SenderChecker{
			AcceptMailFromDomainsOrAddresses(
				[]string{
					"example.org",
					"vodolaz095.ru",
				},
				[]string{
					"a@gmail.com",
					"b@gmail.com",
				}),
		},
	})
	for k, v := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		if err = c.Hello("localhost"); err != nil {
			t.Errorf("HELO failed: %v", err)
		}
		err = c.Mail(k)
		if err != nil {
			if v != nil {
				if err.Error() != v.Error() {
					t.Errorf("wrong error checking %s. Expected: %s. Received: %s", k, v, err)
				}
				continue
			}
			t.Errorf("unexpected error checking %s - %s", k, err)
		}
		err = c.Quit()
		if err != nil {
			t.Errorf("%s : while closing connection", err)
		}
	}
	time.Sleep(time.Second)
	closer()
}
