package connection

import (
	"context"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
)

type testPtrCase struct {
	PTRs  []string
	Error string
}

func TestDenyPTRs(t *testing.T) {
	var denyPtrError = "521 Your IP address is blacklisted. Sorry. You can cry me a river."
	cases := []testPtrCase{
		{[]string{"local"}, ""},
		{[]string{"local", "something.local"}, ""},
		{[]string{"mx.example.org"}, denyPtrError},
		{[]string{"a4-12.smtp-out.eu-west-1.amazonses.com."}, denyPtrError},
		{[]string{"mx.example.org", "a4-12.smtp-out.eu-west-1.amazonses.com."}, denyPtrError},
		{[]string{"a4-12.smtp-out.eu-west-1.amazonses.com.", "mx.example.org"}, denyPtrError},
	}

	for i := range cases {
		addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
			ConnectionCheckers: []msmtpd.ConnectionChecker{
				func(_ context.Context, tr *msmtpd.Transaction) error {
					tr.PTRs = cases[i].PTRs
					return nil
				},
				DenyPTRs([]string{
					"amazonses.com.",
					"mx.example.org",
				}),
			},
		})
		c, err := smtp.Dial(addr)
		if err != nil {
			if err.Error() != cases[i].Error {
				t.Errorf("%s : unexpected error while dialing", err)
			} else {
				t.Logf("error %s is expected", err.Error())
			}
		} else {
			if cases[i].Error != "" {
				t.Errorf("expected error %s is not thrown", cases[i].Error)
			}
			err = c.Close()
			if err != nil {
				t.Errorf("%s : while closing", err)
			}
		}
		closer()
		time.Sleep(time.Second)
	}
}
