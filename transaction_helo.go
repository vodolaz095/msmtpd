package msmptd

import "fmt"

func (t *Transaction) handleHELO(cmd command) {
	var err error
	if len(cmd.fields) < 2 {
		t.reply(502, "i think you have missed parameter")
		t.Hate(missingParameterPenalty)
		return
	}
	if t.HeloName != "" {
		// Reset envelope in case of duplicate HELO
		t.reset()
	}
	t.LogDebug("HELO <%s> is received...", cmd.fields[1])
	for k := range t.server.HeloCheckers {
		err = t.server.HeloCheckers[k](t, cmd.fields[1])
		if err != nil {
			t.error(err)
			return
		}
	}
	t.LogDebug("HELO <%s> is checked!", cmd.fields[1])
	t.mu.Lock()
	t.HeloName = cmd.fields[1]
	t.Protocol = SMTP
	t.mu.Unlock()
	t.reply(250, "Go on, i'm listening...")
	return
}

func (t *Transaction) extensions() []string {
	extensions := []string{
		fmt.Sprintf("SIZE %d", t.server.MaxMessageSize),
		"8BITMIME",
		"PIPELINING",
	}
	if t.server.EnableXCLIENT {
		extensions = append(extensions, "XCLIENT")
	}
	if t.server.TLSConfig != nil && !t.Encrypted {
		extensions = append(extensions, "STARTTLS")
	}
	if t.server.Authenticator != nil && t.Encrypted {
		extensions = append(extensions, "AUTH PLAIN LOGIN")
	}
	return extensions
}

func (t *Transaction) handleEHLO(cmd command) {
	var err error
	if len(cmd.fields) < 2 {
		t.reply(502, "i think you have missed parameter")
		t.Hate(missingParameterPenalty)
		return
	}
	if t.HeloName != "" {
		// Reset envelope in case of duplicate HELO
		t.reset()
	}
	t.LogDebug("EHLO <%s> is received...", cmd.fields[1])
	for k := range t.server.HeloCheckers {
		err = t.server.HeloCheckers[k](t, cmd.fields[1])
		if err != nil {
			t.error(err)
			return
		}
	}
	t.LogDebug("EHLO <%s> is checked!", cmd.fields[1])
	t.mu.Lock()
	t.HeloName = cmd.fields[1]
	t.Protocol = ESMTP
	t.mu.Unlock()
	fmt.Fprintf(t.writer, "250-%s\r\n", t.server.Hostname)
	extensions := t.extensions()
	if len(extensions) > 1 {
		for _, ext := range extensions[:len(extensions)-1] {
			fmt.Fprintf(t.writer, "250-%s\r\n", ext)
		}
	}
	t.reply(250, extensions[len(extensions)-1])
	return
}
