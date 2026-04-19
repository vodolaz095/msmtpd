package msmtpd

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"sync"
	"testing"

	"github.com/vodolaz095/msmtpd/internal"
)

func TestKarma(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		SenderCheckers: []SenderChecker{
			func(_ context.Context, transaction *Transaction) error {
				karma := transaction.Karma()
				if karma != commandExecutedProperly { // because HELO passed
					t.Errorf("wrong initial karma %v", karma)
				}
				if transaction.MailFrom.Address == "scuba@vodolaz095.ru" {
					transaction.Love(1000)
				}
				return nil
			},
		},
		DataHandlers: []DataHandler{
			func(_ context.Context, tr *Transaction) error {
				if tr.Karma() < 1000 {
					t.Errorf("not enough karma. Required at least 1000. Actual: %v", tr.Karma())
				}
				err := ErrorSMTP{
					Code:    555,
					Message: "karma",
				}
				if err.Error() != "555 karma" {
					t.Errorf("wrong error")
				}
				return err
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = c.Hello("mx.example.org")
	if err != nil {
		t.Errorf("sending helo command failed with %s", err)
	}
	err = c.Mail("scuba@vodolaz095.ru")
	if err != nil {
		t.Errorf("sending mail from command failed with %s", err)
	}
	err = c.Rcpt("example@example.org")
	if err != nil {
		t.Errorf("RCPT TO command command failed with %s", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprint(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() != "555 karma" {
			t.Errorf("wrong error returned")
		}
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("sending quit command failed with %s", err)
	}
}

func TestContext(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		ConnectionCheckers: []ConnectionChecker{
			func(_ context.Context, transaction *Transaction) error {
				ctx := transaction.Context()
				t.Logf("context is extracted!")
				go func() {
					t.Logf("starting background goroutine being terminated with context")
					<-ctx.Done()
					wg.Done()
					t.Logf("context is terminated")
				}()
				return nil
			},
		},
	})

	defer closer()
	cm, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	t.Logf("closing connection, so context should be closed too...")
	err = cm.Close()
	if err != nil {
		t.Error(err)
	}
	t.Logf("waiting for context to be closed...")
	wg.Wait()
	t.Logf("context is closed")
}

func TestMeta(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		MaxConnections: 1,
		HeloCheckers: []HelloChecker{
			func(_ context.Context, transaction *Transaction) error {
				name := transaction.HeloName
				transaction.SetFact("something", name)
				transaction.Incr("int64", 1)
				transaction.Incr("float64", 1.1)
				transaction.LogWarn("something")
				transaction.SetFlag("heloCheckerFired")
				return nil
			},
		},
		SenderCheckers: []SenderChecker{
			func(_ context.Context, transaction *Transaction) error {
				var found bool
				_, found = transaction.GetFact("nothing")
				if found {
					t.Error("fact `nothing` is found?")
				}
				something, found := transaction.GetFact("something")
				if !found {
					t.Errorf("fact `something` is not set!")
				}
				if something != "localhost" {
					t.Errorf("wrong meta `something` %s instead of `localhost`", something)
				}
				integerValue, found := transaction.GetCounter("int64")
				if !found {
					t.Errorf("counter `int64` is not set!")
				}
				if integerValue != 1 {
					t.Errorf("wrong value for `int64`")
				}
				floatValue, found := transaction.GetCounter("float64")
				if !found {
					t.Errorf("counter `float64` is not set!")
				}
				if floatValue != 1.1 {
					t.Errorf("wrong value for `float64` - %v", floatValue)
				}
				_, found = transaction.GetCounter("lalala")
				if found {
					t.Errorf("unexistend counter returned value")
				}
				transaction.Incr("int64", 1)
				transaction.Incr("float64", 1.1)
				if !transaction.IsFlagSet("heloCheckerFired") {
					t.Errorf("flag heloCheckerFired is not set")
				}
				transaction.UnsetFlag("heloCheckerFired")
				return nil
			},
		},
		RecipientCheckers: []RecipientChecker{
			func(_ context.Context, transaction *Transaction, _ *mail.Address) error {
				var found bool
				a, found := transaction.GetCounter("int64")
				if !found {
					t.Errorf("counter `int64` is not set!")
				}
				b, found := transaction.GetCounter("float64")
				if !found {
					t.Errorf("counter `float64` is not set!")
				}
				c, found := transaction.GetFact("something")
				if !found {
					t.Errorf("fact `something` is not set!")
				}
				if transaction.IsFlagSet("heloCheckerFired") {
					t.Errorf("flag heloCheckerFired is set")
				}
				return ErrorSMTP{
					Code:    451,
					Message: fmt.Sprintf("%v %v %s", a, b, c),
				}
			},
		},
	})

	defer closer()
	cm, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = cm.Hello("localhost")
	if err != nil {
		t.Error(err)
	}
	err = cm.Mail("somebody@localhost")
	if err != nil {
		t.Error(err)
	}
	err = cm.Rcpt("scuba@example.org")
	if err != nil {
		if err.Error() != "451 2 2.2 localhost" {
			t.Errorf("wrong error `%s` instead `451 2 2.2 localhost`", err)
		}
	}
	err = cm.Close()
	if err != nil {
		t.Error(err)
	}
}
