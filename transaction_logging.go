package msmptd

// LogTrace is used to send trace level message to server logger
func (t *Transaction) LogTrace(format string, args ...any) {
	if t.server != nil {
		if t.server.Logger != nil {
			t.server.Logger.Tracef(t, format, args...)
		}
	}
}

// LogDebug is used to send debug level message to server logger
func (t *Transaction) LogDebug(format string, args ...any) {
	if t.server != nil {
		if t.server.Logger != nil {
			t.server.Logger.Debugf(t, format, args...)
		}
	}
}

// LogInfo is used to send info level message to server logger
func (t *Transaction) LogInfo(format string, args ...any) {
	if t.server != nil {
		if t.server.Logger != nil {
			t.server.Logger.Infof(t, format, args...)
		}
	}
}

// LogWarn is used to send warning level message to server logger
func (t *Transaction) LogWarn(format string, args ...any) {
	if t.server != nil {
		if t.server.Logger != nil {
			t.server.Logger.Warnf(t, format, args...)
		}
	}
}

// LogError is used to send error level message to server logger
func (t *Transaction) LogError(err error, desc string) {
	if t.server != nil {
		if t.server.Logger != nil {
			t.server.Logger.Errorf(t, "%s: %v ", desc, err)
		}
	}
}

// LogFatal is used to send error level message to server logger
func (t *Transaction) LogFatal(err error, desc string) {
	if t.server != nil {
		if t.server.Logger != nil {
			t.server.Logger.Fatalf(t, "%s: %v ", desc, err)
		}
	}
}
