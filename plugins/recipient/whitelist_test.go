package recipient

import (
	"fmt"
	"net/smtp"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

var errRecipientNotWhitelisted = fmt.Errorf("521 I'm sorry, but recipient's email address is not in whitelist")

func TestAcceptMailFromDomainsOrAddresses(t *testing.T) {
	cases := make(map[string]error, 0)

	cases["thisIsNotAEmail"] = fmt.Errorf("502 Malformed e-mail address")

	cases["a@example.org"] = nil
	cases["a@vodolaz095.ru"] = nil
	cases["a@gmail.com"] = nil
	cases["b@gmail.com"] = nil

	cases["d@gmail.com"] = errRecipientNotWhitelisted
	cases["e@gmail.com"] = errRecipientNotWhitelisted
	cases["info@yandex.ru"] = errRecipientNotWhitelisted

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		RecipientCheckers: []msmtpd.RecipientChecker{
			AcceptMailForDomainsOrAddresses(
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
	defer closer()
	for k, v := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		if err = c.Hello("localhost"); err != nil {
			t.Errorf("HELO failed: %v", err)
		}
		err = c.Mail("somebody@example.org")
		if err != nil {
			t.Errorf("MAIL FROM failed: %v", err)
		}
		err = c.Rcpt(k)
		if err != nil {
			if v != nil {
				if err.Error() != v.Error() {
					t.Errorf("wrong error checking %s. Expected: %s. Received: %s", k, v, err)
				}
				continue
			}
			t.Errorf("unexpected error checking %s - %s", k, err)
			continue
		} else {
			if v != nil {
				t.Errorf("error not thrown while checking %s - %s", k, v)
			}
		}
		err = c.Quit()
		if err != nil {
			t.Errorf("%s : while closing connection", err)
		}
	}
}

func TestAcceptMailFromDomains(t *testing.T) {
	cases := make(map[string]error, 0)

	cases["thisIsNotAEmail"] = fmt.Errorf("502 Malformed e-mail address")

	cases["a@example.org"] = nil
	cases["a@vodolaz095.ru"] = nil

	cases["a@gmail.com"] = errRecipientNotWhitelisted
	cases["b@gmail.com"] = errRecipientNotWhitelisted
	cases["d@gmail.com"] = errRecipientNotWhitelisted
	cases["e@gmail.com"] = errRecipientNotWhitelisted
	cases["info@yandex.ru"] = errRecipientNotWhitelisted

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		RecipientCheckers: []msmtpd.RecipientChecker{
			AcceptMailForDomains(
				[]string{
					"example.org",
					"vodolaz095.ru",
				}),
		},
	})
	defer closer()
	for k, v := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		if err = c.Hello("localhost"); err != nil {
			t.Errorf("HELO failed: %v", err)
		}
		err = c.Mail("somebody@example.org")
		if err != nil {
			t.Errorf("MAIL FROM failed: %v", err)
		}
		err = c.Rcpt(k)
		if err != nil {
			if v != nil {
				if err.Error() != v.Error() {
					t.Errorf("wrong error checking %s. Expected: %s. Received: %s", k, v, err)
				}
				continue
			}
			t.Errorf("unexpected error checking %s - %s", k, err)
			continue
		} else {
			if v != nil {
				t.Errorf("error not thrown while checking %s - %s", k, v)
			}
		}
		err = c.Quit()
		if err != nil {
			t.Errorf("%s : while closing connection", err)
		}
	}
}

func TestAcceptMailFromAddresses(t *testing.T) {
	cases := make(map[string]error, 0)

	cases["thisIsNotAEmail"] = fmt.Errorf("502 Malformed e-mail address")

	cases["a@example.org"] = errRecipientNotWhitelisted
	cases["a@vodolaz095.ru"] = errRecipientNotWhitelisted
	cases["a@gmail.com"] = nil
	cases["b@gmail.com"] = nil

	cases["d@gmail.com"] = errRecipientNotWhitelisted
	cases["e@gmail.com"] = errRecipientNotWhitelisted
	cases["info@yandex.ru"] = errRecipientNotWhitelisted

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		RecipientCheckers: []msmtpd.RecipientChecker{
			AcceptMailForAddresses(
				[]string{
					"a@gmail.com",
					"b@gmail.com",
				}),
		},
	})
	defer closer()
	for k, v := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		if err = c.Hello("localhost"); err != nil {
			t.Errorf("HELO failed: %v", err)
		}
		err = c.Mail("somebody@example.org")
		if err != nil {
			t.Errorf("MAIL FROM failed: %v", err)
		}
		err = c.Rcpt(k)
		if err != nil {
			if v != nil {
				if err.Error() != v.Error() {
					t.Errorf("wrong error checking %s. Expected: %s. Received: %s", k, v, err)
				}
				continue
			}
			t.Errorf("unexpected error checking %s - %s", k, err)
			continue
		} else {
			if v != nil {
				t.Errorf("error not thrown while checking %s - %s", k, v)
			}
		}
		err = c.Quit()
		if err != nil {
			t.Errorf("%s : while closing connection", err)
		}
	}
}
