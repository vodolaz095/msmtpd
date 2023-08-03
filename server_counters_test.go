package msmtpd

import (
	"net/smtp"
	"sync"
	"testing"
)

func TestCounters(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	srv := &Server{
		Logger: &TestLogger{Suite: t},
		CloseHandlers: []CloseHandler{
			func(tr *Transaction) error {
				wg.Done()
				return nil
			},
		},
	}
	addr, closer := RunTestServerWithoutTLS(t, srv)
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = c.Hello("localhost")
	if err != nil {
		t.Errorf("%s : while making helo", err)
	}
	t.Logf("Active transactions  - %v", srv.GetActiveTransactionsCount())
	if srv.GetActiveTransactionsCount() != 1 {
		t.Errorf("active transactions - wrong counter %v", srv.GetActiveTransactionsCount())
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while closing", err)
	}
	wg.Wait()
	t.Logf("Transaction is closed")
	if srv.GetActiveTransactionsCount() != 0 {
		t.Errorf("%v active transactions found after close?", srv.GetActiveTransactionsCount())
	}
	t.Logf("Bytes read - %v", srv.GetBytesRead())
	if srv.GetBytesRead() == 0 {
		t.Errorf("bytes read counter not works")
	}
	t.Logf("Bytes wwritten - %v", srv.GetBytesWritten())
	if srv.GetBytesWritten() == 0 {
		t.Errorf("bytes written counter not works")
	}
	t.Logf("Transactions count - %v", srv.GetTransactionsCount())
	if srv.GetTransactionsCount() != 1 {
		t.Errorf("%v finished transaction not counted", srv.GetTransactionsCount())
	}

	t.Logf("Reseting counters")
	srv.ResetCounters()
	t.Logf("Bytes read - %v", srv.GetBytesRead())
	t.Logf("Bytes wwritten - %v", srv.GetBytesWritten())

	t.Logf("Active transactions count - %v", srv.GetActiveTransactionsCount())
	if srv.GetActiveTransactionsCount() != 0 {
		t.Errorf("active transactions found")
	}
	t.Logf("Transactions count - %v", srv.GetTransactionsCount())
	if srv.GetTransactionsCount() != 0 {
		t.Errorf("transaction counter is not reset")
	}
}
