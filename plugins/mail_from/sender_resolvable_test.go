package mail_from

import (
	"fmt"
	"testing"

	"msmtpd"
)

func TestSenderIsResolvableDefault(t *testing.T) {
	testCases := make(map[string]error, 0)

	testCases["info <info@yandex.ru>"] = nil
	testCases["<info@yandex.ru>"] = nil
	testCases["<info@example.org>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)
	testCases["<info@localhost>"] = fmt.Errorf("421 %s", SenderIsNotResolvableComplain)

	handler := SenderIsResolvable(SenderIsResolvableOptions{})
	var err error
	for k, v := range testCases {
		t.Logf("Checking %s...", k)
		tr := msmptd.Transaction{
			ID: fmt.Sprintf("TestSenderIsResolvableDefault - %s", k),
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
