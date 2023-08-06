package msmtpd

import (
	"bytes"
	"io"
	"net/http"
	"net/smtp"
	"sync"
	"testing"
	"time"
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
	if srv.lastTransactionStartedAt.IsZero() {
		t.Errorf("transaction time is not set")
	}
	if time.Now().Sub(srv.lastTransactionStartedAt) > 3*time.Second {
		t.Errorf("last transaction time is too old")
	}
}

func TestCountersHttpExporter(t *testing.T) {
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
	go func() {
		sErr := srv.StartPrometheusScrapperEndpoint("127.0.0.1:5031", "/metrics")
		if sErr != nil {
			t.Fatalf("%s : while starting metrics HTTP server", sErr)
		}
	}()
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

	resp1, err := http.Get("http://127.0.0.1:5031/metrics")
	if err != nil {
		t.Errorf("%s : while getting metrics endpoint", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("wrong status - %s", resp1.Status)
	}
	if resp1.Body != nil {
		defer resp1.Body.Close()
	}
	data, err := io.ReadAll(resp1.Body)
	if err != nil {
		t.Errorf("%s : while reading body", err)
	}
	t.Logf("Body: %s", string(data))
	if !bytes.Contains(data, []byte("bytes_read{hostname=\"localhost.localdomain\"} 22")) {
		t.Errorf("bytes read wrong")
	}
	if !bytes.Contains(data, []byte("bytes_written{hostname=\"localhost.localdomain\"} 199")) {
		t.Errorf("bytes read wrong")
	}
	if !bytes.Contains(data, []byte("active_transactions_count{hostname=\"localhost.localdomain\"} 0")) {
		t.Errorf("bytes read wrong")
	}
	if !bytes.Contains(data, []byte("all_transactions_count{hostname=\"localhost.localdomain\"} 1")) {
		t.Errorf("bytes read wrong")
	}

	t.Logf("Reseting counters")
	srv.ResetCounters()
	resp2, err := http.Get("http://127.0.0.1:5031/metrics")
	if err != nil {
		t.Errorf("%s : while getting metrics endpoint", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("wrong status - %s", resp2.Status)
	}
	if resp2.Body != nil {
		defer resp2.Body.Close()
	}
	data, err = io.ReadAll(resp2.Body)
	if err != nil {
		t.Errorf("%s : while reading body", err)
	}
	if !bytes.Contains(data, []byte("bytes_read{hostname=\"localhost.localdomain\"} 0")) {
		t.Errorf("bytes read wrong")
	}
	if !bytes.Contains(data, []byte("bytes_written{hostname=\"localhost.localdomain\"} 0")) {
		t.Errorf("bytes read wrong")
	}
	if !bytes.Contains(data, []byte("active_transactions_count{hostname=\"localhost.localdomain\"} 0")) {
		t.Errorf("bytes read wrong")
	}
	if !bytes.Contains(data, []byte("all_transactions_count{hostname=\"localhost.localdomain\"} 0")) {
		t.Errorf("bytes read wrong")
	}
	t.Logf("Body: %s", string(data))
}
