package msmtpd

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

func (t *Transaction) serve() {
	defer t.close()
	if !t.server.EnableProxyProtocol {
		t.welcome()
	}
	for {
		for t.scanner.Scan() {
			line := t.scanner.Text()
			t.LogTrace("received: %s", strings.TrimSpace(line))
			t.handle(line)
		}
		err := t.scanner.Err()
		if err == bufio.ErrTooLong {
			t.reply(500, "Line too long")
			// Advance reader to the next newline
			t.reader.ReadString('\n')
			t.scanner = bufio.NewScanner(t.reader)
			// Reset and have the client start over.
			t.reset()
			continue
		}
		break
	}
}

func (t *Transaction) reject() {
	t.reply(421, "I'm tired. Take a break, please.")
	t.close()
}

func (t *Transaction) reset() {
	t.Body = nil
}

func (t *Transaction) welcome() {
	var err error
	for k := range t.server.ConnectionCheckers {
		err = t.server.ConnectionCheckers[k](t)
		if err != nil {
			t.error(err)
			t.close()
			break
		}
	}
	t.reply(220, t.server.WelcomeMessage)
}

func (t *Transaction) reply(code int, message string) {
	t.LogTrace("sending: %d %s", code, message)
	fmt.Fprintf(t.writer, "%d %s\r\n", code, message)
	t.flush()
}

func (t *Transaction) flush() {
	t.conn.SetWriteDeadline(time.Now().Add(t.server.WriteTimeout))
	t.writer.Flush()
	t.conn.SetReadDeadline(time.Now().Add(t.server.ReadTimeout))
}

func (t *Transaction) error(err error) {
	if smtpdError, ok := err.(ErrorSMTP); ok {
		t.reply(smtpdError.Code, smtpdError.Message)
	} else {
		t.reply(502, fmt.Sprintf("%s", err))
	}
}

func (t *Transaction) close() {
	t.LogDebug("Closing transaction...")
	t.writer.Flush()
	time.Sleep(200 * time.Millisecond)
	t.conn.Close()
	t.cancel()
}
