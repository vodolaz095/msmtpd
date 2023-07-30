package helo

import (
	"net"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestDenyReverseDNSMismatch(t *testing.T) {
	cases := []testCase{ //TODO - more and more cases!
		{
			IP:       net.TCPAddr{IP: []byte{193, 41, 76, 25}, Port: 25},
			PTRs:     []string{"r193-41-76-25.utex-telecom.ru.", "r193-41-76-25.utex-telecom.ru"},
			Helo:     "r193-41-76-25.utex-telecom.ru.",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{193, 41, 76, 25}, Port: 25},
			PTRs:     []string{"r193-41-76-25.utex-telecom.ru.", "r193-41-76-25.utex-telecom.ru"},
			Helo:     "r193-41-76-25.utex-telecom.ru",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{193, 41, 76, 25}, Port: 25},
			PTRs:     []string{"r193-41-76-25.utex-telecom.ru."},
			Helo:     "r193-41-76-25.utex-telecom.ru",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{193, 41, 76, 25}, Port: 25},
			PTRs:     []string{"r193-41-76-25.utex-telecom.ru."},
			Helo:     "193.41.76.25",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{213, 180, 204, 89}, Port: 25},
			PTRs:     []string{"204.180.213.in-addr.arpa."},
			Helo:     "204.180.213.in-addr.arpa.",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{185, 245, 187, 136}, Port: 25},
			PTRs:     []string{"136.187.245.185.in-addr.arpa."},
			Helo:     "136.187.245.185.in-addr.arpa.",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{54, 240, 4, 12}, Port: 25},
			PTRs:     nil,
			Helo:     "a4-12.smtp-out.eu-west-1.amazonses.com.",
			ErrorMsg: "",
		},
	}
	heloTestRunner(t, cases, []msmtpd.HelloChecker{DenyReverseDNSMismatch})
}
