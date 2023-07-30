package helo

import (
	"net"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestSkipHeloCheckForLocal(t *testing.T) {
	cases := []testCase{ //TODO - more and more cases!
		{
			IP:   net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 25231},
			Helo: "loopback.local",
			PTRs: []string{
				"something.local",
			},
			ErrorMsg: "",
		},
		{
			IP:   net.TCPAddr{IP: []byte{169, 254, 3, 8}, Port: 25231},
			Helo: "local_unicast.local",
			PTRs: []string{
				"something.local",
			},
			ErrorMsg: "",
		},
		{
			IP:   net.TCPAddr{IP: []byte{192, 168, 1, 3}, Port: 25},
			Helo: "something.local",
			PTRs: []string{
				"something.local",
			},
			ErrorMsg: "",
		},
		{
			IP:   net.TCPAddr{IP: []byte{213, 180, 204, 89}, Port: 25},
			Helo: "something.local",
			PTRs: []string{
				"something.local",
			},
			ErrorMsg: testErrorMessage,
		},
	}

	heloTestRunner(t, cases, []msmtpd.HelloChecker{
		SkipHeloCheckForLocal,
		DenyMalformedDomain,
		DenyBareIP,
		DenyReverseDNSMismatch,
	})
}
