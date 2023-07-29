package msmtpd

import (
	"fmt"
	"log"
)

func (t *Transaction) logEvent(level LoggerLevel, format string, args ...any) {
	var backend Logger
	if t.Logger != nil {
		backend = t.Logger
	} else {
		if t.server != nil {
			if t.server.Logger != nil {
				backend = t.server.Logger
			}
		} else {
			backend = &DefaultLogger{
				Logger: log.Default(),
				Level:  DebugLevel,
			}
		}
	}
	if t.Span != nil {
		t.Span.AddEvent(level.String() + " " + fmt.Sprintf(format, args...))
	}
	switch level {
	case TraceLevel:
		backend.Tracef(t, format, args...)
		break
	case DebugLevel:
		backend.Debugf(t, format, args...)
		break
	case InfoLevel:
		backend.Infof(t, format, args...)
		break
	case WarnLevel:
		backend.Warnf(t, format, args...)
		break
	case ErrorLevel:
		backend.Errorf(t, format, args...)
		break
	case FatalLevel:
		backend.Fatalf(t, format, args...)
		break
	default:
		backend.Infof(t, format, args...)
	}
}

// LogTrace is used to send trace level message to server logger
func (t *Transaction) LogTrace(format string, args ...any) {
	t.logEvent(TraceLevel, format, args...)
}

// LogDebug is used to send debug level message to server logger
func (t *Transaction) LogDebug(format string, args ...any) {
	t.logEvent(DebugLevel, format, args...)
}

// LogInfo is used to send info level message to server logger
func (t *Transaction) LogInfo(format string, args ...any) {
	t.logEvent(InfoLevel, format, args...)
}

// LogWarn is used to send warning level message to server logger
func (t *Transaction) LogWarn(format string, args ...any) {
	t.logEvent(WarnLevel, format, args...)
}

// LogError is used to send error level message to server logger
func (t *Transaction) LogError(err error, desc string) {
	t.logEvent(ErrorLevel, "%s %s", err, desc)
}

// LogFatal is used to send error level message to server logger
func (t *Transaction) LogFatal(err error, desc string) {
	t.logEvent(FatalLevel, "%s %s", err, desc)
}
