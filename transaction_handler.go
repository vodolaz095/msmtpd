package msmptd

func (t *Transaction) handle(line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.LogTrace("Command received: %s", line)
	cmd := parseLine(line)

	// Commands are dispatched to the appropriate handler functions.
	// If a network error occurs during handling, the handler should
	// just return and let the error be handled on the next read.

	switch cmd.action {
	case "PROXY":
		t.handlePROXY(cmd)
		return
	case "HELO":
		t.handleHELO(cmd)
		return
	case "EHLO":
		t.handleEHLO(cmd)
		return
	case "MAIL":
		t.handleMAIL(cmd)
		return
	case "RCPT":
		t.handleRCPT(cmd)
		return
	case "STARTTLS":
		t.handleSTARTTLS(cmd)
		return
	case "DATA":
		t.handleDATA(cmd)
		return
	case "RSET":
		t.handleRSET(cmd)
		return
	case "NOOP":
		t.handleNOOP(cmd)
		return
	case "QUIT":
		t.handleQUIT(cmd)
		return
	case "AUTH":
		t.handleAUTH(cmd)
		return
	case "XCLIENT":
		t.handleXCLIENT(cmd)
		return
	}
	t.Hate(unknownCommandPenalty)
	t.LogDebug("Unsupported command received: %s", line)
	t.reply(502, "Unsupported command.")
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
	t.reply(221, "Farewell, my friend!")
	t.close()
	return
}
