package msmptd

import (
	"bytes"
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	buffer := bytes.NewBufferString("")
	backend := log.New(buffer, "", log.Lshortfile)
	testLogger := DefaultLogger{
		Logger: backend,
		Level:  DebugLevel,
	}
	tr := &Transaction{
		ID: "testTransaction1",
	}
	testLogger.Tracef(tr, "Tracef %s", "trace")
	testLogger.Debugf(tr, "Debugf %s", "debug")
	testLogger.Infof(tr, "Infof %s", "info")
	testLogger.Warnf(tr, "Warnf %s", "warn")
	testLogger.Errorf(tr, "Errorf %s", "error")
	t.Logf("Logged: %s", buffer.String())
}
