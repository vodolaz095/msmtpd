package msmtpd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"sync"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd/internal"
)

func TestSMTP(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if supported, _ := c.Extension("AUTH"); supported {
		t.Error("AUTH supported before TLS")
	}
	if supported, _ := c.Extension("8BITMIME"); !supported {
		t.Error("8BITMIME not supported")
	}
	if supported, _ := c.Extension("STARTTLS"); supported {
		t.Error("STARTTLS supported")
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	err = c.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}

	err = c.Verify("foobar@example.net")
	if err == nil {
		t.Error("Unexpected support for VRFY")
	}

	if err = internal.DoCommand(c.Text, 250, "NOOP"); err != nil {
		t.Errorf("NOOP failed: %v", err)
	}

	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestListenAndServe(t *testing.T) {
	server := &Server{}
	addr, closer := RunTestServerWithoutTLS(t, server)
	closer()
	go func() {
		lsErr := server.ListenAndServe(addr)
		if lsErr != nil {
			t.Errorf("%s : while starting server on %s", lsErr, server.Address())
		}
	}()
	time.Sleep(100 * time.Millisecond)
	if server.Address().String() != addr {
		t.Errorf("server is listening on `%s` instead of `%s",
			server.Address(), addr,
		)
	}
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestMaxConnections(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Logger:         &TestLogger{Suite: t},
		MaxConnections: 1,
	})
	defer closer()
	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	_, err = smtp.Dial(addr)
	if err == nil {
		t.Error("Dial succeeded despite MaxConnections = 1")
	}
	c1.Close()
}

func TestNoMaxConnections(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Logger:         &TestLogger{Suite: t},
		MaxConnections: -1,
	})
	defer closer()
	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	c1.Close()
}

func TestTimeoutClose(t *testing.T) {
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Logger:         &TestLogger{Suite: t},
		MaxConnections: 1,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	})
	defer closer()
	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	time.Sleep(time.Second * 2)
	c2, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c1.Mail("sender@example.org"); err == nil {
		t.Error("MAIL succeeded despite being timed out.")
	}
	if err = c2.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c2.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
	c2.Close()
}

func TestTLSTimeout(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
		Logger:       &TestLogger{Suite: t},
		ReadTimeout:  time.Second * 2,
		WriteTimeout: time.Second * 2,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	time.Sleep(time.Second)
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestCloseHandlers(t *testing.T) {
	wg := sync.WaitGroup{}
	var closeHandler1Called bool
	var closeHandler2Called bool
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Logger: &TestLogger{Suite: t},
		ConnectionCheckers: []ConnectionChecker{
			func(transaction *Transaction) error {
				t.Logf("Giving 2 wg to transaction %s", transaction.ID)
				wg.Add(2)
				return nil
			},
		},
		CloseHandlers: []CloseHandler{
			func(transaction *Transaction) error {
				t.Logf("Closing transaction %s by 1st handler", transaction.ID)
				closeHandler1Called = true
				wg.Done()
				return nil
			},
			func(transaction *Transaction) error {
				t.Logf("Closing transaction %s by 2nd handler", transaction.ID)
				closeHandler2Called = true
				wg.Done()
				return fmt.Errorf("who cares")
			},
		},
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	err = c.Close()
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	wg.Wait()
	if !closeHandler1Called {
		t.Errorf("close handler 1 is not called")
	}
	if !closeHandler2Called {
		t.Errorf("close handler 2 is not called")
	}
}

func TestTLSListener(t *testing.T) {
	cert, err := tls.X509KeyPair(internal.LocalhostCert, internal.LocalhostKey)
	if err != nil {
		t.Errorf("Cert load failed: %v", err)
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	defer ln.Close()
	addr := ln.Addr().String()
	server := &Server{
		Logger: &TestLogger{Suite: t},
		Authenticator: func(tr *Transaction, username, password string) error {
			if tr.TLS == nil {
				t.Error("didn't correctly set connection state on TLS connection")
			}
			return nil
		},
	}
	go func() {
		server.Serve(ln)
	}()
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Errorf("couldn't connect to tls socket: %v", err)
	}
	c, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		t.Errorf("couldn't create client: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 334, "AUTH PLAIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = internal.DoCommand(c.Text, 235, "Zm9vAGJhcgBxdXV4"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestShutdown(t *testing.T) {
	t.Logf("Starting shutdown test")
	server := &Server{
		Logger: &TestLogger{Suite: t},
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("Listen failed: %v", err)
	}
	srvres := make(chan error)
	go func() {
		t.Log("Starting server")
		srvres <- server.Serve(ln)
	}()
	// Connect a client
	c, err := smtp.Dial(ln.Addr().String())
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	// While the client connection is open, shut down the server (without
	// waiting for it to finish)
	err = server.Shutdown(false)
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
	// Verify that Shutdown() worked by attempting to connect another client
	_, err = smtp.Dial(ln.Addr().String())
	if err == nil {
		t.Errorf("Dial did not fail as expected")
	}
	if _, typok := err.(*net.OpError); !typok {
		t.Errorf("Dial did not return net.OpError as expected: %v (%T)", err, err)
	}
	// Wait for shutdown to complete
	shutres := make(chan error)
	go func() {
		t.Log("Waiting for server shutdown to finish")
		shutres <- server.Wait()
	}()
	// Slight delay to ensure Shutdown() blocks
	time.Sleep(250 * time.Millisecond)
	// Wait() should not have returned yet due to open client conn
	select {
	case shuterr := <-shutres:
		t.Errorf("Wait() returned early w/ error: %v", shuterr)
	default:
	}
	// Now close the client
	t.Log("Closing client connection")
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
	c.Close()

	// Wait for Wait() to return
	t.Log("Waiting for Wait() to return")
	select {
	case shuterr := <-shutres:
		if shuterr != nil {
			t.Errorf("Wait() returned error: %v", shuterr)
		}
	case <-time.After(15 * time.Second):
		t.Errorf("Timed out waiting for Wait() to return")
	}

	// Wait for Serve() to return
	t.Log("Waiting for Serve() to return")
	select {
	case srverr := <-srvres:
		if srverr != ErrServerClosed {
			t.Errorf("Serve() returned error: %v", srverr)
		}
	case <-time.After(15 * time.Second):
		t.Errorf("Timed out waiting for Serve() to return")
	}
}

func TestServeFailsIfShutdown(t *testing.T) {
	server := &Server{
		Logger: &TestLogger{Suite: t},
	}
	err := server.Shutdown(true)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
	err = server.Serve(nil)
	if err != ErrServerClosed {
		t.Errorf("Serve() did not return ErrServerClosed: %v", err)
	}
}

func TestWaitFailsIfNotShutdown(t *testing.T) {
	server := &Server{
		Logger: &TestLogger{Suite: t},
	}
	err := server.Wait()
	if err == nil {
		t.Errorf("Wait() did not fail as expected")
	}
}
