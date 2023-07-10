package msmtpd

import (
	"fmt"
	"log"
)

// Logger is interface all Server loggers must satisfy
type Logger interface {
	Tracef(transaction *Transaction, format string, args ...any)
	Debugf(transaction *Transaction, format string, args ...any)
	Infof(transaction *Transaction, format string, args ...any)
	Warnf(transaction *Transaction, format string, args ...any)
	Errorf(transaction *Transaction, format string, args ...any)
	Fatalf(transaction *Transaction, format string, args ...any)
}

// LoggerLevel describes logging level like JournalD has by
// https://github.com/coreos/go-systemd/blob/main/journal/journal.go
// See for inspiration https://stackoverflow.com/questions/2031163/when-to-use-the-different-log-levels
type LoggerLevel uint8

// TraceLevel is used when we record very verbose message like SMTP protocol raw data being sent/received
const TraceLevel LoggerLevel = 8

// DebugLevel is used when we log information that is diagnostically helpful to people more than just developers (IT, sysadmins, etc.).
const DebugLevel LoggerLevel = 7

// InfoLevel is used when we log generally useful information to log
// (service start/stop, configuration assumptions, etc).
// This is information we always want to be available but we usually don't care about
// under normal circumstances. This is my out-of-the-box config level.
const InfoLevel LoggerLevel = 6

// WarnLevel is used when we log anything that can potentially cause application oddities,
// but for which we are automatically recovering.
// Such as switching from a primary to backup server, retrying an operation, missing secondary data, etc.
const WarnLevel LoggerLevel = 4

// ErrorLevel is used for any error which is fatal to the operation, but not the service or application (can't open a required file, missing data, etc.).
// These errors will force user (administrator, or direct user) intervention.
const ErrorLevel LoggerLevel = 3

// FatalLevel is used for any error that is forcing a shutdown of the service or
// application to prevent data loss (or further data loss).
const FatalLevel LoggerLevel = 2

// DefaultLogger is logger by default using standard library logger as backend https://pkg.go.dev/log
type DefaultLogger struct {
	*log.Logger
	Level LoggerLevel
}

// Tracef sends TraceLevel message
func (d *DefaultLogger) Tracef(transaction *Transaction, format string, args ...any) {
	if d.Level >= TraceLevel {
		d.Printf("TRACE [%s]: %s", transaction.ID, fmt.Sprintf(format, args...))
	}
}

// Debugf sends DebugLevel message
func (d *DefaultLogger) Debugf(transaction *Transaction, format string, args ...any) {
	if d.Level >= DebugLevel {
		d.Printf("DEBUG [%s]: %s", transaction.ID, fmt.Sprintf(format, args...))
	}
}

// Infof sends InfoLevel message
func (d *DefaultLogger) Infof(transaction *Transaction, format string, args ...any) {
	if d.Level >= InfoLevel {
		d.Printf("INFO [%s]: %s", transaction.ID, fmt.Sprintf(format, args...))
	}
}

// Warnf sends WarnLevel message
func (d *DefaultLogger) Warnf(transaction *Transaction, format string, args ...any) {
	if d.Level >= WarnLevel {
		d.Printf("WARN [%s]: %s", transaction.ID, fmt.Sprintf(format, args...))
	}
}

// Errorf sends ErrorLevel message
func (d *DefaultLogger) Errorf(transaction *Transaction, format string, args ...any) {
	if d.Level >= ErrorLevel {
		d.Printf("ERROR [%s]: %s", transaction.ID, fmt.Sprintf(format, args...))
	}
}

// Fatalf sends FatalLevel message and stops application with exit code 1
func (d *DefaultLogger) Fatalf(transaction *Transaction, format string, args ...any) {
	d.Logger.Fatalf("FATAL [%s]: %s", transaction.ID, fmt.Sprintf(format, args...))
}
