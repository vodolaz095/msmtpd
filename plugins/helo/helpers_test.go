package helo

import (
	"fmt"
	"net"
	"net/textproto"
	"testing"

	"msmtpd"
)

type testLogger struct{}

func (tl *testLogger) Tracef(transaction *msmtpd.Transaction, format string, args ...any) {
	fmt.Printf("TRACE: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Debugf(transaction *msmtpd.Transaction, format string, args ...any) {
	fmt.Printf("DEBUG: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Infof(transaction *msmtpd.Transaction, format string, args ...any) {
	fmt.Printf("INFO: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Warnf(transaction *msmtpd.Transaction, format string, args ...any) {
	fmt.Printf("WARN: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Errorf(transaction *msmtpd.Transaction, format string, args ...any) {
	fmt.Printf("ERROR: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Fatalf(transaction *msmtpd.Transaction, format string, args ...any) {
	panic("it is bad")
}

func runserver(t *testing.T, server *msmtpd.Server) (addr string, closer func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("Listen failed: %v", err)
	}
	logger := testLogger{}
	server.Logger = &logger
	go func() {
		serveErr := server.Serve(ln)
		if err != nil {
			t.Errorf("%s : while starting server on %s",
				serveErr, server.Address())
		}
	}()
	done := make(chan bool)
	go func() {
		<-done
		ln.Close()
	}()
	return ln.Addr().String(), func() {
		done <- true
	}
}

func cmd(c *textproto.Conn, expectedCode int, format string, args ...interface{}) error {
	id, err := c.Cmd(format, args...)
	if err != nil {
		return err
	}
	c.StartResponse(id)
	_, _, err = c.ReadResponse(expectedCode)
	c.EndResponse(id)
	return err
}
