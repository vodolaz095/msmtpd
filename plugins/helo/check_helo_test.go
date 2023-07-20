package helo

import (
	"fmt"
	"net"
	"net/smtp"
	"testing"

	"msmtpd"
)

var errGeneralComplain = fmt.Errorf("521 %s", complain)

type testCase struct {
	HELO  string
	Error error
}

func TestCheckHELO_Dynamic(t *testing.T) {
	cases := []testCase{
		{
			HELO:  "Sodom",
			Error: errGeneralComplain,
		},
		{
			HELO:  "193.41.76.125",
			Error: errGeneralComplain,
		},
		{
			HELO:  "R193-41-76-125.utex-telecom.ru",
			Error: nil, // important!
		},
		{
			HELO:  "mail.example.org",
			Error: errGeneralComplain,
		},
	}
	addr, closer := runserver(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			func(transaction *msmtpd.Transaction) error {
				transaction.Addr = &net.TCPAddr{
					IP:   net.ParseIP("193.41.76.125"),
					Port: 49762,
				}
				return nil
			},
		},
		HeloCheckers: []msmtpd.HelloChecker{
			CheckHELO(Options{
				TolerateInvalidHostname: false,
				TolerateBareIP:          false,
				TolerateDynamic:         true, // so i can actually ran this test with familiar for me IP address
				TolerateRDNSMismatch:    false,
			}),
		},
	})
	defer closer()
	for k := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		err = c.Hello(cases[k].HELO)
		if err != nil {
			if cases[k].Error != nil {
				if err.Error() == cases[k].Error.Error() {
					t.Logf("%s - FAILED AS EXPECTED", cases[k].HELO)
				} else {
					t.Errorf("wrong error checking - %s. Expected: %s. Received: %s",
						cases[k].HELO, cases[k].Error, err)

				}
			} else {
				t.Errorf("unexpected error checking >>>%s<<< - %s", cases[k].HELO, err)
			}
		} else {
			if cases[k].Error != nil {
				t.Errorf("expected error `%s` was not thrown while checking >>>%s<<< - ",
					cases[k].Error, cases[k].HELO)
			}
		}
		c.Quit() // it is ok
	}
}

func TestCheckHELO_Default(t *testing.T) {
	cases := []testCase{
		{
			HELO:  "Sodom",
			Error: errGeneralComplain,
		},
		{
			HELO:  "193.41.76.125",
			Error: errGeneralComplain,
		},
		{
			HELO:  "R193-41-76-125.utex-telecom.ru",
			Error: errGeneralComplain,
		},
		{
			HELO:  "mail.example.org",
			Error: errGeneralComplain,
		},
	}
	addr, closer := runserver(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			func(transaction *msmtpd.Transaction) error {
				transaction.Addr = &net.TCPAddr{
					IP:   net.ParseIP("193.41.76.125"),
					Port: 49762,
				}
				return nil
			},
		},
		HeloCheckers: []msmtpd.HelloChecker{
			CheckHELO(Options{
				TolerateInvalidHostname: false,
				TolerateBareIP:          false,
				TolerateDynamic:         false,
				TolerateRDNSMismatch:    false,
			}),
		},
	})
	defer closer()
	for k := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		err = c.Hello(cases[k].HELO)
		if err != nil {
			if cases[k].Error != nil {
				if err.Error() == cases[k].Error.Error() {
					t.Logf("%s - FAILED AS EXPECTED", cases[k].HELO)
				} else {
					t.Errorf("wrong error checking - %s. Expected: %s. Received: %s",
						cases[k].HELO, cases[k].Error, err)

				}
			} else {
				t.Errorf("unexpected error checking >>>%s<<< - %s", cases[k].HELO, err)
			}
		} else {
			if cases[k].Error != nil {
				t.Errorf("expected error `%s` was not thrown while checking >>>%s<<< - ",
					cases[k].Error, cases[k].HELO)
			}
		}
		c.Quit() // it is ok
	}
}

func TestCheckHELO_IgnoreHostnameForLocalAddresses(t *testing.T) {
	cases := []testCase{
		{
			HELO:  "localhost",
			Error: nil, // because we ran from local IP address :-)
		},
		{
			HELO:  "localhost.localdomain",
			Error: nil, // because we ran from local IP address :-)
		},
		{
			HELO:  "something.local",
			Error: nil, // because we ran from local IP address :-)
		},
		{
			HELO:  "mail.example.org",
			Error: nil, // because we ran from local IP address :-)
		},
	}
	addr, closer := runserver(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			func(transaction *msmtpd.Transaction) error {
				transaction.Addr = &net.TCPAddr{
					IP:   net.ParseIP("192.168.1.2"),
					Port: 49762,
				}
				return nil
			},
		},
		HeloCheckers: []msmtpd.HelloChecker{
			CheckHELO(Options{
				TolerateInvalidHostname:         false,
				TolerateBareIP:                  false,
				TolerateDynamic:                 false,
				TolerateRDNSMismatch:            false,
				IgnoreHostnameForLocalAddresses: true,
			}),
		},
	})
	defer closer()
	for k := range cases {
		c, err := smtp.Dial(addr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
		}
		err = c.Hello(cases[k].HELO)
		if err != nil {
			if cases[k].Error != nil {
				if err.Error() == cases[k].Error.Error() {
					t.Logf("%s - FAILED AS EXPECTED", cases[k].HELO)
				} else {
					t.Errorf("wrong error checking - %s. Expected: %s. Received: %s",
						cases[k].HELO, cases[k].Error, err)

				}
			} else {
				t.Errorf("unexpected error checking >>>%s<<< - %s", cases[k].HELO, err)
			}
		} else {
			if cases[k].Error != nil {
				t.Errorf("expected error `%s` was not thrown while checking >>>%s<<< - ",
					cases[k].Error, cases[k].HELO)
			}
		}
		c.Quit() // it is ok
	}
}
