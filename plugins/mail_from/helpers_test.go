package mail_from

import (
	"fmt"
	"net"
	"testing"

	"msmtpd"
)

type testLogger struct{}

func (tl *testLogger) Tracef(transaction *msmptd.Transaction, format string, args ...any) {
	fmt.Printf("TRACE: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Debugf(transaction *msmptd.Transaction, format string, args ...any) {
	fmt.Printf("DEBUG: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Infof(transaction *msmptd.Transaction, format string, args ...any) {
	fmt.Printf("INFO: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Warnf(transaction *msmptd.Transaction, format string, args ...any) {
	fmt.Printf("WARN: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Errorf(transaction *msmptd.Transaction, format string, args ...any) {
	fmt.Printf("ERROR: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Fatalf(transaction *msmptd.Transaction, format string, args ...any) {
	panic("it is bad")
}

func runserver(t *testing.T, server *msmptd.Server) (addr string, closer func()) {
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
