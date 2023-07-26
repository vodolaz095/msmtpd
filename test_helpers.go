package msmtpd

import (
	"fmt"
	"net"
	"testing"

	"msmtpd/internal"
)

type TestLogger struct {
	Suite *testing.T
}

func (tl *TestLogger) Tracef(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("TRACE: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Debugf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("DEBUG: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Infof(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("INFO: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Warnf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("WARN: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Errorf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("ERROR: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Fatalf(transaction *Transaction, format string, args ...any) {
	tl.Suite.Logf("FATAL: [%s] %s %s\n",
		tl.Suite.Name(), transaction.ID, fmt.Sprintf(format, args...))
	tl.Suite.Errorf(format, args...)
}

func AuthenticatorForTestsThatAlwaysWorks(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and succeed!", username, password)
	return nil
}

func AuthenticatorForTestsThatAlwaysFails(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and fail!", username, password)
	return ErrorSMTP{Code: 550, Message: "Denied"}
}

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

func RunTestServerWithTLS(t *testing.T, server *Server) (addr string, closer func()) {
	cfg, err := internal.MakeTLSForLocalhost()
	if err != nil {
		t.Fatalf("%s : while loading test certs for localhost", err)
	}
	server.TLSConfig = cfg
	return RunTestServerWithoutTLS(t, server)
}
