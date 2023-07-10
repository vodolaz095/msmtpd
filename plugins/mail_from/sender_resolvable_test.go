package mail_from

import (
	"fmt"
	"testing"

	"msmtpd"
)

func TestSenderIsResolvableDefault(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["thisIsNotAEmail"] = fmt.Errorf("521 %s", "I cannot parse your address, it seems malformed. i'm sorry.")

	testCases["info <info@yandex.ru>"] = nil
	testCases["info <info@mx.yandex.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)
	testCases["<info@yandex.ru>"] = nil
	testCases["<info@example.org>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)
	testCases["<info@localhost>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	// it should fail, becase A/AAAA Fallback delivery is disabled from the box
	testCases["info <info@mx.yandex.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	// dramatic misuse of cloudflare :-)
	// feedback.vodolaz095.ru.	33	IN	MX	10 ivory.vodolaz095.ru.
	// ivory.vodolaz095.ru.	4	IN	A	192.168.1.2
	// it should fail, because 192.168.1.2 is local IP
	testCases["<somebody@feedback.vodolaz095.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)
	testCases["<somebody@ivory.vodolaz095.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	handler := SenderIsResolvable(SenderIsResolvableOptions{})
	var err error
	for k, v := range testCases {
		t.Logf("Checking %s...", k)
		tr := msmptd.Transaction{
			ID: fmt.Sprintf("TestSenderIsResolvableDefault - %s", k),
		}
		err = handler(&tr, k)
		if err != nil {
			if v != nil {
				t.Logf("Checking %s - unexpected error %s", k, err)
				if v.Error() != err.Error() {
					t.Errorf("for testing %s wrong error was thrown : %s", k, err)
				}
			} else {
				t.Errorf("for testing %s unexpected error was thrown : %s", k, err)
			}
		}
	}
}

func TestSenderIsResolvableFallback(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["thisIsNotAEmail"] = fmt.Errorf("521 %s", "I cannot parse your address, it seems malformed. i'm sorry.")

	testCases["info <info@yandex.ru>"] = nil
	testCases["<info@yandex.ru>"] = nil
	testCases["<info@example.org>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)
	testCases["<info@localhost>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	// it should work according to standards (https://en.wikipedia.org/wiki/MX_record#Fallback_to_the_address_record)
	// because mx.yandex.ru has A record and 25th port open for connections
	// but yandex will refuse mail like this
	testCases["info <info@mx.yandex.ru>"] = nil

	// dramatic misuse of cloudflare :-)
	// feedback.vodolaz095.ru.	33	IN	MX	10 ivory.vodolaz095.ru.
	// ivory.vodolaz095.ru.	4	IN	A	192.168.1.2
	// it should fail, because 192.168.1.2 is local IP
	testCases["<somebody@feedback.vodolaz095.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)
	testCases["<somebody@ivory.vodolaz095.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	handler := SenderIsResolvable(SenderIsResolvableOptions{
		FallbackToAddressRecord: true,
	})
	var err error
	for k, v := range testCases {
		t.Logf("Checking %s...", k)
		tr := msmptd.Transaction{
			ID: fmt.Sprintf("TestSenderIsResolvableFallback - %s", k),
		}
		err = handler(&tr, k)
		if err != nil {
			t.Logf("Checking %s - error %s", k, err)
			if v != nil {
				if v.Error() != err.Error() {
					t.Errorf("for testing %s wrong error was thrown : %s", k, err)
				}
			} else {
				t.Errorf("for testing %s unexpected error was thrown : %s", k, err)
			}
		}
	}
}

func TestSenderIsResolvableLocal(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["thisIsNotAEmail"] = fmt.Errorf("521 %s", "I cannot parse your address, it seems malformed. i'm sorry.")

	testCases["info <info@yandex.ru>"] = nil
	testCases["<info@yandex.ru>"] = nil
	testCases["<info@example.org>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)
	testCases["<info@localhost>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	// this should fail, because we disabled A/AAAA record fallback delivery
	testCases["info <info@mx.yandex.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	// dramatic misuse of cloudflare :-)
	// feedback.vodolaz095.ru.	33	IN	MX	10 ivory.vodolaz095.ru.
	// ivory.vodolaz095.ru.	4	IN	A	192.168.1.2
	// it should work, because we enabled delivery to local addresses
	testCases["<somebody@feedback.vodolaz095.ru>"] = nil
	// but this should fail, we disabled A/AAAA record fallback delivery
	testCases["<somebody@ivory.vodolaz095.ru>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	handler := SenderIsResolvable(SenderIsResolvableOptions{
		FallbackToAddressRecord: false,
		AllowLocalAddresses:     true,
	})
	var err error
	for k, v := range testCases {
		t.Logf("Checking %s...", k)
		tr := msmptd.Transaction{
			ID: fmt.Sprintf("TestSenderIsResolvableFallback - %s", k),
		}
		err = handler(&tr, k)
		if err != nil {
			t.Logf("Checking %s - error %s", k, err)
			if v != nil {
				if v.Error() != err.Error() {
					t.Errorf("for testing %s wrong error was thrown : %s", k, err)
				}
			} else {
				t.Errorf("for testing %s unexpected error was thrown : %s", k, err)
			}
		}
	}
}

func TestSenderIsResolvableFallbackAndLocal(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["info <info@yandex.ru>"] = nil
	testCases["<info@yandex.ru>"] = nil
	testCases["<info@example.org>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	// it should work according to standards :-)
	testCases["info <info@mx.yandex.ru>"] = nil
	// providing local loop back as MX server is usually used to troll spammers
	testCases["<info@localhost>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	// dramatic misuse of cloudflare :-)
	testCases["<somebody@feedback.vodolaz095.ru>"] = nil
	testCases["<somebody@ivory.vodolaz095.ru>"] = nil

	handler := SenderIsResolvable(SenderIsResolvableOptions{
		FallbackToAddressRecord: true,
		AllowLocalAddresses:     true,
	})
	var err error
	for k, v := range testCases {
		t.Logf("Checking %s...", k)
		tr := msmptd.Transaction{
			ID: fmt.Sprintf("TestSenderIsResolvableFallback - FROM: `%s`", k),
		}
		err = handler(&tr, k)
		if err != nil {
			t.Logf("Checking %s - error %s", k, err)
			if v != nil {
				if v.Error() != err.Error() {
					t.Errorf("for testing %s wrong error was thrown : %s", k, err)
				}
			} else {
				t.Errorf("for testing %s unexpected error was thrown : %s", k, err)
			}
		}
	}
}
