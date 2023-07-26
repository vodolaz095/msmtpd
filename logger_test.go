package msmtpd

import (
	"bytes"
	"log"
	"testing"
)

func TestTransactionLogger(t *testing.T) {
	buffer := bytes.NewBufferString("")
	backend := log.New(buffer, "", log.Lshortfile)
	testLoggerForThisTest := DefaultLogger{
		Logger: backend,
		Level:  TraceLevel,
	}
	tr := &Transaction{
		ID: "testTransaction1",
	}
	testLoggerForThisTest.Tracef(tr, "Tracef %s", "trace")
	testLoggerForThisTest.Debugf(tr, "Debugf %s", "debug")
	testLoggerForThisTest.Infof(tr, "Infof %s", "info")
	testLoggerForThisTest.Warnf(tr, "Warnf %s", "warn")
	testLoggerForThisTest.Errorf(tr, "Errorf %s", "error")
	t.Logf("Logged: %s", buffer.String())
}
