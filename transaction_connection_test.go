package msmtpd

import (
	"errors"
	"net/smtp"
	"testing"
)

func TestConnectionCheck(t *testing.T) {
	cc := make([]ConnectionChecker, 0)
	cc = append(cc, func(tr *Transaction) error {
		return ErrorSMTP{Code: 552, Message: "Denied"}
	})
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		ConnectionCheckers: cc,
	})
	defer closer()
	if _, err := smtp.Dial(addr); err == nil {
		t.Error("Dial succeeded despite ConnectionCheck")
	}
}

func TestConnectionCheckSimpleError(t *testing.T) {
	cc := make([]ConnectionChecker, 0)
	cc = append(cc, func(tr *Transaction) error {
		return errors.New("Denied")
	})
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		ConnectionCheckers: cc,
	})
	defer closer()
	if _, err := smtp.Dial(addr); err == nil {
		t.Error("Dial succeeded despite ConnectionCheck")
	}
}
