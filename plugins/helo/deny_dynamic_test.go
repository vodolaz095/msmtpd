package helo

import (
	"net"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestDenyDynamicIP(t *testing.T) {
	cases := []testCase{ //TODO - more and more cases!
		{
			IP:       net.TCPAddr{IP: []byte{193, 41, 76, 25}, Port: 25},
			PTRs:     []string{"r193-41-76-25.utex-telecom.ru."},
			Helo:     "r193-41-76-25.utex-telecom.ru.",
			ErrorMsg: testErrorMessage,
		},
		{
			IP:       net.TCPAddr{IP: []byte{193, 41, 76, 25}, Port: 25},
			PTRs:     []string{"r193-41-76-25.utex-telecom.ru."},
			Helo:     "193.41.76.25",
			ErrorMsg: testErrorMessage,
		},
		{
			IP:       net.TCPAddr{IP: []byte{213, 180, 204, 89}, Port: 25},
			PTRs:     []string{"204.180.213.in-addr.arpa."},
			Helo:     "204.180.213.in-addr.arpa.",
			ErrorMsg: testErrorMessage,
		},
		{
			IP:       net.TCPAddr{IP: []byte{185, 245, 187, 136}, Port: 25},
			PTRs:     []string{"136.187.245.185.in-addr.arpa."},
			Helo:     "136.187.245.185.in-addr.arpa.",
			ErrorMsg: testErrorMessage,
		},
		{
			IP:       net.TCPAddr{IP: []byte{54, 240, 4, 12}, Port: 25},
			PTRs:     []string{"a4-12.smtp-out.eu-west-1.amazonses.com."},
			Helo:     "a4-12.smtp-out.eu-west-1.amazonses.com.",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{122, 160, 68, 26}, Port: 25},
			PTRs:     []string{"abts-north-static-026.68.160.122.airtelbroadband.in."},
			Helo:     "abts-north-static-026.68.160.122.airtelbroadband.in",
			ErrorMsg: testErrorMessage,
		},
	}
	heloTestRunner(t, cases, []msmtpd.HelloChecker{DenyDynamicIP})
}
