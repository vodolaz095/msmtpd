package msmtpd

import "fmt"

func (t *Transaction) handle(line string) {
	t.LogDebug("Command received: %s", line)
	cmd := parseLine(line)
	// Commands are dispatched to the appropriate handler functions.
	// If a network error occurs during handling, the handler should
	// just return and let the error be handled on the next read.

	switch cmd.action {
	case "PROXY":
		t.handlePROXY(cmd)
	case "HELO":
		t.handleHELO(cmd)
	case "EHLO":
		t.handleEHLO(cmd)
	case "MAIL":
		t.handleMAIL(cmd)
	case "RCPT":
		t.handleRCPT(cmd)
	case "STARTTLS":
		t.handleSTARTTLS(cmd)
	case "DATA":
		t.handleDATA(cmd)
	case "RSET":
		t.handleRSET(cmd)
	case "NOOP":
		t.handleNOOP(cmd)
	case "QUIT":
		t.handleQUIT(cmd)
	case "AUTH":
		t.handleAUTH(cmd)
	case "XCLIENT":
		t.handleXCLIENT(cmd)
	default:
		t.Hate(unknownCommandPenalty)
		t.LogDebug("Unsupported command received: %s", line)
		t.reply(502, "Unsupported command.")
	}
}

func (t *Transaction) handleRSET(_ command) {
	t.reset()
	t.reply(250, "I forgot everything you have said, go ahead please!")
}

func (t *Transaction) handleNOOP(_ command) {
	t.reply(250, "I'm finishing procrastinating, go ahead please!")
}

func (t *Transaction) handleQUIT(_ command) {
	t.reply(221, fmt.Sprintf("Farewell, my friend! Transaction %s is finished", t.ID))
	t.close()
}
