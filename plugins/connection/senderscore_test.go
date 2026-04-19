package connection

import (
	"context"
	"net"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
)

func TestCheckSenderScore(t *testing.T) {
	cases := []net.TCPAddr{
		{IP: []byte{193, 41, 76, 25}, Port: 25},
		{IP: []byte{193, 41, 76, 171}, Port: 25},
		{IP: []byte{54, 211, 7, 12}, Port: 25},
		{IP: []byte{54, 145, 43, 133}, Port: 25},
		{IP: []byte{193, 176, 233, 203}, Port: 25},
		{IP: []byte{94, 156, 102, 151}, Port: 25},
		{IP: []byte{94, 156, 102, 151}},
	}

	for i := range cases {
		addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
			ConnectionCheckers: []msmtpd.ConnectionChecker{
				func(_ context.Context, tr *msmtpd.Transaction) error {
					tr.Addr = &cases[i]
					return nil
				},
				RequireSenderScore(1),
				func(_ context.Context, tr *msmtpd.Transaction) error {
					senderscore, found := tr.GetCounter(SenderscoreCounter)
					if !found {
						t.Errorf("senderscore not found for %s", tr.Addr.String())
					}
					t.Logf("IP: %s. Senderscore: %.0f", tr.Addr.String(), senderscore)
					return nil
				},
			},
		})
		c, err := smtp.Dial(addr)
		if err != nil {
			if err.Error() != "421 Service not available. Try again later, please." {
				t.Errorf("%s : unexpected error while dialing", err)
			}
		} else {
			err = c.Close()
			if err != nil {
				t.Errorf("%s : while closing", err)
			}
		}
		closer()
		time.Sleep(time.Second)
	}
}
