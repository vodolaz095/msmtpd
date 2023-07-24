package karma

import (
	"bytes"
	"fmt"
	"net"
	"net/textproto"
	"testing"
	"time"

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

func MakeTestMessage(from, to string) string {
	now := time.Now()
	buh := bytes.NewBufferString("Date: " + now.Format(time.RFC1123Z) + "\r\n")
	buh.WriteString("To: " + to + "\r\n")
	buh.WriteString("From: " + from + "\r\n")
	buh.WriteString(fmt.Sprintf("Subject: Test email send on %s\r\n", now.Format(time.RFC1123Z)))
	buh.WriteString(fmt.Sprintf("Message-Id: %s@localhost\r\n", now.Format("20060102150405")))
	buh.WriteString("\r\n")
	buh.WriteString(fmt.Sprintf("This is test message send from %s to %s on %s\r\n",
		from, to, now.Format(time.Stamp),
	))
	return buh.String()
}
