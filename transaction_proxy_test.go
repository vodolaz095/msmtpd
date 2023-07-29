package msmtpd

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"testing"

	"github.com/vodolaz095/msmtpd/internal"
)

func TestProxyNotEnabled(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		EnableProxyProtocol: false, // important
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}

	where := strings.Split(addr, ":")
	err = internal.DoCommand(c.Text, 550, "PROXY TCP4 8.8.8.8 %s 443 %s", where[0], where[1])
	if err != nil {
		t.Errorf("sending proxy command enabled from the box - %s", err)
	}

	err = c.Hello("nobody.example.org")
	if err != nil {
		t.Errorf("sending helo command failed with %s", err)
	}

	err = c.Quit()
	if err != nil {
		t.Errorf("sending quit command failed with %s", err)
	}
}

func TestProxyEnabledSuccess(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		EnableProxyProtocol: true, // important
		ConnectionCheckers: []ConnectionChecker{
			func(tr *Transaction) error {
				tr.LogInfo("Remote address for connection checker is %s", tr.Addr)
				tr.SetFact("originalRemote", tr.Addr.String())
				return nil
			},
		},
		HeloCheckers: []HelloChecker{
			func(tr *Transaction) error {
				tr.LogInfo("Remote address for HELO is %s", tr.Addr)
				if tr.Addr.String() != "8.8.8.8:443" {
					t.Errorf("remote address is not set properly %s", tr.Addr)
				}
				original, found := tr.GetFact("originalRemote")
				if !found {
					t.Errorf("original address is not set")
				}
				if original == tr.Addr.String() {
					t.Errorf("address not changed")
				}
				return nil
			},
		},
	})
	defer closer()

	con, err := net.Dial("tcp", addr)
	if err != nil {
		t.Errorf("%s: error dialing %s", err, addr)
	}
	where := strings.Split(addr, ":")
	_, err = fmt.Fprintf(con, "PROXY TCP4 8.8.8.8 %s 443 %s\r\n", where[0], where[1])
	if err != nil {
		t.Errorf("%s : while sending proxy command", err)
		return
	}
	c, err := smtp.NewClient(con, "localhost")
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}

	err = c.Hello("nobody.example.org")
	if err != nil {
		t.Errorf("sending helo command failed with %s", err)
	}

	err = c.Quit()
	if err != nil {
		t.Errorf("sending quit command failed with %s", err)
	}
}

func TestProxyEnabledMalformedProtocol(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		EnableProxyProtocol: true, // important
	})
	defer closer()

	con, err := net.Dial("tcp", addr)
	if err != nil {
		t.Errorf("%s: error dialing %s", err, addr)
	}
	where := strings.Split(addr, ":")
	_, err = fmt.Fprintf(con, "PROXY UDP 8.8.8.8 %s port %s\r\n", where[0], where[1])
	if err != nil {
		t.Errorf("%s : while sending proxy command", err)
		return
	}
	_, err = smtp.NewClient(con, "localhost")
	if err != nil {
		if err.Error() == "502 unable to decode proxy protocol - only TCP4/TCP6 is supported" {
			t.Logf("proxy command failed with malformed protocol")
		} else {
			t.Errorf("%s : unexpected error", err)
		}
	} else {
		t.Errorf("proxy error not thrown for malformed protocol")
	}
}

func TestProxyEnabledMalformedPort(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		EnableProxyProtocol: true, // important
	})
	defer closer()

	con, err := net.Dial("tcp", addr)
	if err != nil {
		t.Errorf("%s: error dialing %s", err, addr)
	}
	where := strings.Split(addr, ":")
	_, err = fmt.Fprintf(con, "PROXY TCP4 8.8.8.8 %s port %s\r\n", where[0], where[1])
	if err != nil {
		t.Errorf("%s : while sending proxy command", err)
		return
	}
	_, err = smtp.NewClient(con, "localhost")
	if err != nil {
		if err.Error() == "502 malformed port in proxy command" {
			t.Logf("proxy command failed with malformed port")
		} else {
			t.Errorf("%s : unexpected error", err)
		}
	} else {
		t.Errorf("proxy error not thrown for malformed port")
	}
}

func TestProxyEnabledMalformedAddress(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		EnableProxyProtocol: true, // important
	})
	defer closer()

	con, err := net.Dial("tcp", addr)
	if err != nil {
		t.Errorf("%s: error dialing %s", err, addr)
	}
	where := strings.Split(addr, ":")
	_, err = fmt.Fprintf(con, "PROXY TCP4 339.69.72.11 %s 443 %s\r\n", where[0], where[1])
	if err != nil {
		t.Errorf("%s : while sending proxy command", err)
		return
	}
	_, err = smtp.NewClient(con, "localhost")
	if err != nil {
		if err.Error() == "502 malformed network address" {
			t.Logf("proxy command failed with malformed address")
		} else {
			t.Errorf("%s : unexpected error", err)
		}
	} else {
		t.Errorf("proxy error not thrown for malformed address")
	}
}

func TestProxyEnabledMalformedManyArguments(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		EnableProxyProtocol: true, // important
	})
	defer closer()

	con, err := net.Dial("tcp", addr)
	if err != nil {
		t.Errorf("%s: error dialing %s", err, addr)
	}
	where := strings.Split(addr, ":")
	_, err = fmt.Fprintf(con, "PROXY TCP4 %s %s\r\n", where[0], where[1])
	if err != nil {
		t.Errorf("%s : while sending proxy command", err)
		return
	}
	_, err = smtp.NewClient(con, "localhost")
	if err != nil {
		if err.Error() == "502 malformed proxy command" {
			t.Logf("proxy command failed with malformed address")
		} else {
			t.Errorf("%s : unexpected error", err)
		}
	} else {
		t.Errorf("proxy error not thrown for malformed address")
	}
}
