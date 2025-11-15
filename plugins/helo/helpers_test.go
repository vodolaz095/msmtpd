package helo

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
)

type testCase struct {
	IP       net.TCPAddr
	PTRs     []string
	Helo     string
	ErrorMsg string
}

func (tc *testCase) String() string {
	return fmt.Sprintf("CASE: IP=%s PTRs=%v HELO=%s ERR=%s",
		tc.IP.String(),
		tc.PTRs,
		tc.Helo,
		tc.ErrorMsg,
	)
}

func heloTestRunner(t *testing.T, cases []testCase, checkers []msmtpd.HelloChecker) {
	for k := range cases {
		addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
			ConnectionCheckers: []msmtpd.ConnectionChecker{
				func(_ context.Context, tr *msmtpd.Transaction) error {
					tr.Addr = &cases[k].IP
					//					tr.PTRs = []string{cases[k].Helo}
					if len(cases[k].PTRs) > 0 {
						tr.PTRs = cases[k].PTRs
					}
					return nil
				},
			},
			HeloCheckers: checkers,
		})

		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v for %v", err, k)
		}
		err = c.Hello(cases[k].Helo)
		if cases[k].ErrorMsg == "" {
			if err != nil {
				t.Errorf("FAIL: %s : unexpected error for testing %v - %s", err, k, cases[k].String())
			}
		} else {
			if err != nil {
				if err.Error() != cases[k].ErrorMsg {
					t.Errorf("FAIL: wrong error: %s vs %s", err.Error(), cases[k].ErrorMsg)
				} else {
					t.Logf("error %s thrown as expected for %v", err.Error(), cases[k])
				}
			} else {
				t.Errorf("FAIL: expected error %s is not thrown for %v - %s", cases[k].ErrorMsg, k, cases[k].String())
			}
		}
		err = c.Close()
		if err != nil {
			t.Errorf("FAIL: %s : while closing connection for %v %s", err, k, cases[k].String())
		}
		time.Sleep(100 * time.Millisecond)
		closer()
	}
}
