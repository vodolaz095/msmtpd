package connection

import (
	"net/smtp"
	"testing"

	"msmtpd"
)

func TestWhitelistFail(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Whitelist([]string{"192.168.1.3"}),
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("Dial failed with wrong error: %s", err)
		}
		return
	}
	t.Errorf("error is not thrown")
}

func TestWhitelistPass(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Whitelist([]string{"127.0.0.1"}),
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("%s : while dialing", err)
		return
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while quiting", err)
		return
	}
}

func TestWhitelistPassSubnet1(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Whitelist([]string{"127.0.0"}),
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("%s : while dialing", err)
		return
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while quiting", err)
		return
	}
}

func TestWhitelistPassSubnet2(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Whitelist([]string{"127.0"}),
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("%s : while dialing", err)
		return
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while quiting", err)
		return
	}
}

func TestWhitelistPassSubnet3(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Whitelist([]string{"127"}),
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("%s : while dialing", err)
		return
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while quiting", err)
		return
	}
}

func TestBlacklistFail(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Blacklist([]string{"127.0.0.1"}),
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("Dial failed with wrong error: %s", err)
		}
		return
	}
	t.Errorf("error is not thrown")
}

func TestBlacklistFailSubnet1(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Blacklist([]string{"127.0.0"}),
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("Dial failed with wrong error: %s", err)
		}
		return
	}
	t.Errorf("error is not thrown")
}

func TestBlacklistFailSubnet2(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Blacklist([]string{"127.0"}),
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("Dial failed with wrong error: %s", err)
		}
		return
	}
	t.Errorf("error is not thrown")
}

func TestBlacklistFailSubnet3(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Blacklist([]string{"127.0"}),
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("Dial failed with wrong error: %s", err)
		}
		return
	}
	t.Errorf("error is not thrown")
}

func TestBlacklistFailSubnet4(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Blacklist([]string{"127"}),
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("Dial failed with wrong error: %s", err)
		}
		return
	}
	t.Errorf("error is not thrown")
}

func TestBlacklistSuccess(t *testing.T) {
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			Blacklist([]string{"8.8.8.8"}),
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("%s : while dialing", err)
		return
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while quiting", err)
		return
	}
}
