package sender

import (
	"fmt"
	"net/smtp"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestSenderIsResolvableDefault(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["info@yandex.ru"] = nil
	testCases["info@mx.yandex.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)
	testCases["info@yandex.ru"] = nil
	testCases["info@example.org"] = fmt.Errorf("421 %s", IsNotResolvableComplain)
	testCases["info@localhost"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	// it should fail, becase A/AAAA Fallback delivery is disabled from the box
	testCases["info@mx.yandex.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	// dramatic misuse of cloudflare :-)
	// feedback.vodolaz095.ru.	33	IN	MX	10 ivory.vodolaz095.ru.
	// ivory.vodolaz095.ru.	4	IN	A	192.168.1.2
	// it should fail, because 192.168.1.2 is local IP
	testCases["somebody@feedback.vodolaz095.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)
	testCases["somebody@ivory.vodolaz095.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		SenderCheckers: []msmtpd.SenderChecker{
			IsResolvable(IsResolvableOptions{}),
		},
	})
	defer closer()

	for k, v := range testCases {
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
}

func TestSenderIsResolvableFallback(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["info@yandex.ru"] = nil
	testCases["info@example.org"] = fmt.Errorf("421 %s", IsNotResolvableComplain)
	testCases["info@localhost"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	// it should work according to standards (https://en.wikipedia.org/wiki/MX_record#Fallback_to_the_address_record)
	// because mx.yandex.ru has A record and 25th port open for connections
	// but yandex will refuse mail like this
	testCases["info@mx.yandex.ru"] = nil

	// dramatic misuse of cloudflare :-)
	// feedback.vodolaz095.ru.	33	IN	MX	10 ivory.vodolaz095.ru.
	// ivory.vodolaz095.ru.	4	IN	A	192.168.1.2
	// it should fail, because 192.168.1.2 is local IP
	testCases["somebody@feedback.vodolaz095.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)
	testCases["somebody@ivory.vodolaz095.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		SenderCheckers: []msmtpd.SenderChecker{
			IsResolvable(IsResolvableOptions{
				FallbackToAddressRecord: true,
			}),
		},
	})
	defer closer()

	for k, v := range testCases {
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
}

func TestSenderIsResolvableLocal(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["info@yandex.ru"] = nil
	testCases["info@example.org"] = fmt.Errorf("421 %s", IsNotResolvableComplain)
	testCases["info@localhost"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	// this should fail, because we disabled A/AAAA record fallback delivery
	testCases["info@mx.yandex.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	// dramatic misuse of cloudflare :-)
	// feedback.vodolaz095.ru.	33	IN	MX	10 ivory.vodolaz095.ru.
	// ivory.vodolaz095.ru.	4	IN	A	192.168.1.2
	// it should work, because we enabled delivery to local addresses
	testCases["somebody@feedback.vodolaz095.ru"] = nil
	// but this should fail, we disabled A/AAAA record fallback delivery
	testCases["somebody@ivory.vodolaz095.ru"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		SenderCheckers: []msmtpd.SenderChecker{
			IsResolvable(IsResolvableOptions{
				FallbackToAddressRecord: false,
				AllowLocalAddresses:     true,
			}),
		},
	})
	defer closer()

	for k, v := range testCases {
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
}

func TestSenderIsResolvableFallbackAndLocal(t *testing.T) {
	testCases := make(map[string]error, 0)
	testCases["info@yandex.ru"] = nil
	testCases["info@example.org"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	// it should work according to standards :-)
	testCases["info@mx.yandex.ru"] = nil
	// providing local loop back as MX server is usually used to troll spammers
	testCases["info@localhost"] = fmt.Errorf("421 %s", IsNotResolvableComplain)

	// dramatic misuse of cloudflare :-)
	testCases["somebody@feedback.vodolaz095.ru"] = nil
	testCases["somebody@ivory.vodolaz095.ru"] = nil

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		SenderCheckers: []msmtpd.SenderChecker{
			IsResolvable(IsResolvableOptions{
				FallbackToAddressRecord: true,
				AllowLocalAddresses:     true,
			}),
		},
	})
	defer closer()

	for k, v := range testCases {
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
}
