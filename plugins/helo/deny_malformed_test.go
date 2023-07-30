package helo

import (
	"testing"

	"github.com/vodolaz095/msmtpd"
)

const testErrorMessage = "521 I don't like the way you introduce yourself. Goodbye!"

func TestDenyMalformed(t *testing.T) {
	cases := []testCase{ //TODO - more and more cases!
		{
			Helo:     "localhost",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "something.oldcity",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "sodom",
			ErrorMsg: testErrorMessage,
		},
		{
			Helo:     "136.187.245.185.in-addr.arpa.",
			ErrorMsg: "",
		},
		{
			Helo:     "a4-12.smtp-out.eu-west-1.amazonses.com.",
			ErrorMsg: "",
		},
		{
			Helo:     "mail.ru",
			ErrorMsg: "",
		},
	}
	heloTestRunner(t, cases, []msmtpd.HelloChecker{DenyMalformedDomain})
}
