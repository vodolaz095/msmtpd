package msmptd

import (
	"net/mail"
	"strings"
)

func (t *Transaction) handleMAIL(cmd command) {
	if len(cmd.params) != 2 || strings.ToUpper(cmd.params[0]) != "FROM" {
		t.Hate(missingParameterPenalty)
		t.reply(502, "Invalid syntax.")
		return
	}
	if t.HeloName == "" {
		t.Hate(missingParameterPenalty)
		t.reply(502, "Please introduce yourself first.")
		return
	}
	if t.server.Authenticator != nil && t.Username == "" {
		t.Hate(missingParameterPenalty)
		t.reply(530, "Authentication Required.")
		return
	}
	if !t.Encrypted && t.server.ForceTLS {
		t.Hate(missingParameterPenalty)
		t.reply(502, "Please turn on TLS by issuing a STARTTLS command.")
		return
	}
	if t.MailFrom.Address != "" {
		t.Hate(missingParameterPenalty)
		t.reply(502, "Duplicate MAIL")
		return
	}
	var err error
	var addr *mail.Address // null sender
	// We must accept a null sender as per rfc5321 section-6.1.
	if cmd.params[1] != "<>" {
		addr, err = parseAddress(cmd.params[1])
		if err != nil {
			t.reply(502, "Malformed e-mail address")
			return
		}
	}
	t.LogDebug("MAIL FROM [%s] is received...", cmd.fields[1])
	for k := range t.server.SenderCheckers {
		err = t.server.SenderCheckers[k](t, addr.Address)
		if err != nil {
			t.error(err)
			return
		}
	}
	t.LogDebug("MAIL FROM [%s] is checked...", cmd.fields[1])
	t.MailFrom = *addr
	t.reply(250, "Ok, it makes sense, go ahead please!")
	return
}
