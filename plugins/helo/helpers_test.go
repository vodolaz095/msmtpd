package helo

import (
	"fmt"
	"testing"

	"msmtpd"
	"msmtpd/internal"
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
	logger := testLogger{}
	server.Logger = &logger
	return internal.RunServerWithoutTLS(t, server)
}
