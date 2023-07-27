package msmtpd

import (
	"fmt"
	"net"
	"testing"

	"msmtpd/internal"
)

// TestLogger is logger being used only in unit tests
type TestLogger struct {
	Suite *testing.T
}

// Tracef records trace level message
func (tl *TestLogger) Tracef(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("TRACE: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

// Debugf records debug level message
func (tl *TestLogger) Debugf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("DEBUG: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

// Infof records info level message
func (tl *TestLogger) Infof(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("INFO: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

// Warnf records warning level message
func (tl *TestLogger) Warnf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("WARN: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

// Errorf records error level message
func (tl *TestLogger) Errorf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("ERROR: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

// Fatalf records fatal level message and fails test
func (tl *TestLogger) Fatalf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("FATAL: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
	tl.Suite.Errorf(format, args...)
}

// AuthenticatorForTestsThatAlwaysWorks should not be used for production
func AuthenticatorForTestsThatAlwaysWorks(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and succeed!", username, password)
	return nil
}

// AuthenticatorForTestsThatAlwaysFails should not be used for production
func AuthenticatorForTestsThatAlwaysFails(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and fail!", username, password)
	return ErrorSMTP{Code: 550, Message: "Denied"}
}

// RunTestServerWithoutTLS runs test server for unit tests without TLS support
func RunTestServerWithoutTLS(t *testing.T, server *Server) (addr string, closer func()) {
	logger := TestLogger{Suite: t}
	server.Logger = &logger
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	go func() {
		server.Serve(ln)
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

// RunTestServerWithTLS runs test server for unit tests with TLS support and cert for localhost
func RunTestServerWithTLS(t *testing.T, server *Server) (addr string, closer func()) {
	cfg, err := internal.MakeTLSForLocalhost()
	if err != nil {
		t.Fatalf("%s : while loading test certs for localhost", err)
	}
	server.TLSConfig = cfg
	return RunTestServerWithoutTLS(t, server)
}
