package msmtpd

import (
	"errors"
	"net/smtp"
	"sync"
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

func TestConnectionCheckerRejectingAndCloseHandler(t *testing.T) {
	var connectionHandlerCalled, closeHandlerCalled bool
	wg := sync.WaitGroup{}
	wg.Add(2)
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		ConnectionCheckers: []ConnectionChecker{
			func(tr *Transaction) error {
				connectionHandlerCalled = true
				wg.Done()
				return ErrorSMTP{Code: 521, Message: "i do not like you"}
			},
		},
		CloseHandlers: []CloseHandler{
			func(tr *Transaction) error {
				t.Logf("close handler is called")
				closeHandlerCalled = true
				wg.Done()
				return nil
			},
		},
	})
	defer closer()
	if _, err := smtp.Dial(addr); err == nil {
		t.Error("Dial succeeded despite ConnectionCheck")
	}
	wg.Wait()
	if !connectionHandlerCalled {
		t.Error("connection handler not called")
	}
	if !closeHandlerCalled {
		t.Error("close handler not called")
	}

}
