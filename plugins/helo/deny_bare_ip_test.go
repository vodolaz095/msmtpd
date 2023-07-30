package helo

import (
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestDenyBareIP(t *testing.T) {
	cases := []testCase{ //TODO - more and more cases!
		{
			Helo:     "127.0.0.1",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "192.168.1.3",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "192.168.1.3",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "8.8.8.8",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "2a0e:e5c3:5651:dc3e:bba6:618:7f63:cabf",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "mail.ru",
			ErrorMsg: "",
		},
		{
			Helo:     "localhost",
			ErrorMsg: "",
		},
	}

	heloTestRunner(t, cases, []msmtpd.HelloChecker{DenyBareIP})
}
