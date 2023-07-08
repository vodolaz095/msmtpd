package msmptd

// LogTrace is used to send trace level message to server logger
func (t *Transaction) LogTrace(format string, args ...any) {
	t.server.Logger.Tracef(t, format, args...)
}

// LogDebug is used to send debug level message to server logger
func (t *Transaction) LogDebug(format string, args ...any) {
	t.server.Logger.Debugf(t, format, args...)
}

// LogInfo is used to send info level message to server logger
func (t *Transaction) LogInfo(format string, args ...any) {
	t.server.Logger.Infof(t, format, args...)
}

// LogWarn is used to send warning level message to server logger
func (t *Transaction) LogWarn(format string, args ...any) {
	t.server.Logger.Warnf(t, format, args...)
}

// LogError is used to send error level message to server logger
func (t *Transaction) LogError(err error, desc string) {
	t.server.Logger.Errorf(t, "%s: %v ", desc, err)
}
