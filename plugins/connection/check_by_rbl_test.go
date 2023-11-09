package connection

import (
	"context"
	"net"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
)

func TestCheckByRbl(t *testing.T) {
	cases := []net.TCPAddr{
		{IP: []byte{193, 41, 76, 25}, Port: 25},
		{IP: []byte{193, 41, 76, 171}, Port: 25},
		{IP: []byte{54, 211, 7, 12}, Port: 25},
		{IP: []byte{54, 145, 43, 133}, Port: 25},
		{IP: []byte{193, 176, 233, 203}, Port: 25},
	}
	var lists []string

	lists = append(lists, SpamhauseReverseIPBlackLists...)
	lists = append(lists, SpamEatingMonkeyReverseIPBlackLists...)
	lists = append(lists, SorbsReverseIPBlacklists...)
	lists = append(lists, SpamratsIPBlacklists...)

	for i := range cases {
		addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{}
					return d.DialContext(ctx, network, "8.8.8.8:53")
				},
			},
			ConnectionCheckers: []msmtpd.ConnectionChecker{
				func(tr *msmtpd.Transaction) error {
					tr.Addr = &cases[i]
					return nil
				},
				CheckByReverseIPBlacklists(1, lists),
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
