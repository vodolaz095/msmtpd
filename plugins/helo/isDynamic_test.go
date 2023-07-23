package helo

import (
	"fmt"
	"net"
	"testing"
	"time"

	"msmtpd"
)

type isDtc struct {
	IP      net.TCPAddr
	Helo    string
	Dynamic bool
}

func TestIsDynamic(t *testing.T) {
	var tr msmtpd.Transaction
	var res bool
	cases := []isDtc{ //TODO - more and more cases!
		{net.TCPAddr{IP: []byte{193, 41, 76, 25}, Port: 25},
			"r193-41-76-25.utex-telecom.ru.", true},
		{net.TCPAddr{IP: []byte{213, 180, 204, 89}, Port: 25},
			"204.180.213.in-addr.arpa.", true},
		{net.TCPAddr{IP: []byte{185, 245, 187, 136}, Port: 25},
			"136.187.245.185.in-addr.arpa.", true},
		{net.TCPAddr{IP: []byte{185, 245, 187, 136}, Port: 25},
			"136.187.245.185.in-addr.arpa.", true},
		{net.TCPAddr{IP: []byte{54, 240, 4, 12}, Port: 25},
			"a4-12.smtp-out.eu-west-1.amazonses.com.", false,
		},
	}
	logger := testLogger{}
	for i := range cases {
		tr = msmtpd.Transaction{
			ID: fmt.Sprintf("transaction %v %s %s", i,
				cases[i].IP.String(), cases[i].Helo),
			StartedAt:  time.Now(),
			ServerName: "",
			HeloName:   cases[i].Helo,
			Addr:       &cases[i].IP,
			Logger:     &logger,
		}
		res = isDynamic(&tr)
		if res != cases[i].Dynamic {
			t.Errorf("case %v %s %s. Status %v, while it should be %v",
				i, cases[i].IP.String(), cases[i].Helo, res, cases[i].Dynamic)
		} else {
			t.Logf("Test passed for case %v %s %s",
				i, cases[i].IP.String(), cases[i].Helo)
		}

	}

}
