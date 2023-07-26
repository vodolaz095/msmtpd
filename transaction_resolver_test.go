package msmtpd

import (
	"context"
	"net"
	"net/smtp"
	"testing"
	"time"
)

func TestTransaction_Resolver(t *testing.T) {
	tr := Transaction{
		ID: "testTransactionResolver",
	}
	resolver := tr.Resolver()
	addrs, err := resolver.LookupMX(context.Background(), "yandex.ru")
	if err != nil {
		t.Errorf("%s : while resolving MX for yandex.ru", err)
	}
	for i := range addrs {
		t.Logf("MX %s %v", addrs[i].Host, addrs[i].Pref)
	}
}

func TestTransaction_Resolver_In_Server(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		HeloCheckers: []HelloChecker{
			func(tr *Transaction) error {
				name := tr.HeloName
				addrs, err := tr.Resolver().LookupMX(context.Background(), name)
				if err != nil {
					return err
				}
				if name != "yandex.ru" {
					t.Errorf("wrong HELO provided %s", name)
				}
				for i := range addrs {
					t.Logf("MX %s %v resolved for %s", addrs[i].Host, addrs[i].Pref, name)
				}
				if len(addrs) != 1 {
					t.Errorf("not enough mx resolved")
					return ErrorSMTP{Code: 555, Message: "not enough mx resolved"}
				}
				if addrs[0].Host != "mx.yandex.ru." {
					t.Errorf("wrong mx resolved - %s", addrs[0].Host)
				}
				if addrs[0].Pref != 10 {
					t.Errorf("wrong mx priority - %v", addrs[0].Pref)
				}
				return nil
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("yandex.ru"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestTransaction_Resolver_In_Server_Custom(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(10000),
				}
				return d.DialContext(ctx, network, "1.1.1.1:53")
			},
		},
		HeloCheckers: []HelloChecker{
			func(tr *Transaction) error {
				name := tr.HeloName
				addrs, err := tr.Resolver().LookupMX(context.Background(), name)
				if err != nil {
					return err
				}
				if name != "yandex.ru" {
					t.Errorf("wrong HELO provided %s", name)
				}
				for i := range addrs {
					t.Logf("MX %s %v resolved for %s", addrs[i].Host, addrs[i].Pref, name)
				}
				if len(addrs) != 1 {
					t.Errorf("not enough mx resolved")
					return ErrorSMTP{Code: 555, Message: "not enough mx resolved"}
				}
				if addrs[0].Host != "mx.yandex.ru." {
					t.Errorf("wrong mx resolved - %s", addrs[0].Host)
				}
				if addrs[0].Pref != 10 {
					t.Errorf("wrong mx priority - %v", addrs[0].Pref)
				}
				return nil
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("yandex.ru"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}
