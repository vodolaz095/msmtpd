package msmtpd

import "fmt"

func (t *Transaction) handle(line string) {
	t.LogTrace("Command received: %s", line)
	cmd := parseLine(line)
	// Commands are dispatched to the appropriate handler functions.
	// If a network error occurs during handling, the handler should
	// just return and let the error be handled on the next read.

	switch cmd.action {
	case "PROXY":
		t.handlePROXY(cmd)
		break
	case "HELO":
		t.handleHELO(cmd)
		break
	case "EHLO":
		t.handleEHLO(cmd)
		break
	case "MAIL":
		t.handleMAIL(cmd)
		break
	case "RCPT":
		t.handleRCPT(cmd)
		break
	case "STARTTLS":
		t.handleSTARTTLS(cmd)
		break
	case "DATA":
		t.handleDATA(cmd)
		break
	case "RSET":
		t.handleRSET(cmd)
		break
	case "NOOP":
		t.handleNOOP(cmd)
		break
	case "QUIT":
		t.handleQUIT(cmd)
		break
	case "AUTH":
		t.handleAUTH(cmd)
		break
	case "XCLIENT":
		t.handleXCLIENT(cmd)
		break
	default:
		t.Hate(unknownCommandPenalty)
		t.LogDebug("Unsupported command received: %s", line)
		t.reply(502, "Unsupported command.")
	}
}

func (t *Transaction) handleRSET(cmd command) {
	t.reset()
	t.reply(250, "I forgot everything you have said, go ahead please!")
	return
}

func (t *Transaction) handleNOOP(cmd command) {
	t.reply(250, "I'm finishing procrastinating, go ahead please!")
	return
}

func (t *Transaction) handleQUIT(cmd command) {
	t.reply(221, fmt.Sprintf("Farewell, my friend! Transaction %s is finished", t.ID))
	t.close()
	return
}
