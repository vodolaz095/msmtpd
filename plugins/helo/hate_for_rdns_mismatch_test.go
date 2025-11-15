package helo

import (
	"context"
	"net"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestHateForRDNSMismatch(t *testing.T) {
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
			IP:       net.TCPAddr{IP: []byte{91, 239, 5, 18}, Port: 25},
			PTRs:     nil,
			Helo:     "mail.astralnalog.ru.",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{91, 239, 5, 18}, Port: 25},
			PTRs:     nil,
			Helo:     "mail.astral.ru.",
			ErrorMsg: "",
		},
	}
	heloTestRunner(t, cases, []msmtpd.HelloChecker{
		HateForRDNSMismatch(10),
		func(_ context.Context, tr *msmtpd.Transaction) error {
			karma := tr.Karma()
			t.Logf("For HELO %s karma is %v", tr.HeloName, karma)
			if karma != -10 {
				t.Errorf("For helo %s negative karma not applied", tr.HeloName)
			}
			return nil
		},
	})
}
