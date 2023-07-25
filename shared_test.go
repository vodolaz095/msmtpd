package msmtpd

import (
	"fmt"
	"testing"

	"msmtpd/internal"
)

type testLogger struct{}

func (tl *testLogger) Tracef(transaction *Transaction, format string, args ...any) {
	fmt.Printf("TRACE: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Debugf(transaction *Transaction, format string, args ...any) {
	fmt.Printf("DEBUG: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Infof(transaction *Transaction, format string, args ...any) {
	fmt.Printf("INFO: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Warnf(transaction *Transaction, format string, args ...any) {
	fmt.Printf("WARN: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Errorf(transaction *Transaction, format string, args ...any) {
	fmt.Printf("ERROR: %s %s\n", transaction.ID, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Fatalf(transaction *Transaction, format string, args ...any) {
	panic("it is bad")
}

func AuthenticatorForTestsThatAlwaysWorks(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and succeed!", username, password)
	return nil
}

func AuthenticatorForTestsThatAlwaysFails(tr *Transaction, username, password string) error {
	tr.LogInfo("Pretend we authenticate as %s %s and fail!", username, password)
	return ErrorSMTP{Code: 550, Message: "Denied"}
}

func RunServerWithoutTLS(t *testing.T, server *Server) (addr string, closer func()) {
	logger := testLogger{}
	server.Logger = &logger
	return internal.RunServerWithoutTLS(t, server)
}

func RunServerWithTLS(t *testing.T, server *Server) (addr string, closer func()) {
	cfg, err := internal.MakeTLSForLocalhost()
	if err != nil {
		t.Fatalf("%s : while loading test certs for localhost", err)
	}
	server.TLSConfig = cfg
	return RunServerWithoutTLS(t, server)
}
