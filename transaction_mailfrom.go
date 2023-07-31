package msmtpd

import (
	"net/mail"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

func (t *Transaction) handleMAIL(cmd command) {
	if len(cmd.params) != 2 || strings.ToUpper(cmd.params[0]) != "FROM" {
		t.Hate(missingParameterPenalty)
		t.reply(502, "Invalid syntax.")
		return
	}
	if t.HeloName == "" {
		t.Hate(missingParameterPenalty)
		t.LogDebug("MAIL FROM called without HELO/EHLO")
		t.reply(502, "Please introduce yourself first.")
		return
	}
	if !t.Encrypted && t.server.ForceTLS {
		t.LogDebug("MAIL FROM called without STARTTLS")
		t.Hate(missingParameterPenalty)
		t.reply(502, "Please turn on TLS by issuing a STARTTLS command.")
		return
	}
	if t.server.Authenticator != nil && t.Username == "" {
		t.LogDebug("MAIL FROM called without authentication")
		t.Hate(missingParameterPenalty)
		t.reply(530, "Authentication Required.")
		return
	}
	if t.MailFrom.Address != "" {
		t.LogDebug("MAIL FROM was already called")
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
		t.MailFrom = *addr
	} else {
		t.MailFrom = mail.Address{}
	}
	t.LogDebug("Checking MAIL FROM %s by %v SenderCheckers...",
		t.MailFrom.String(), len(t.server.SenderCheckers),
	)
	t.Span.SetAttributes(attribute.String("from", t.MailFrom.String()))
	for k := range t.server.SenderCheckers {
		err = t.server.SenderCheckers[k](t)
		if err != nil {
			t.error(err)
			return
		}
	}
	t.LogInfo("MAIL FROM %s is checked by %v SenderCheckers and accepted!",
		t.MailFrom.String(), len(t.server.SenderCheckers),
	)
	t.reply(250, "Ok, it makes sense, go ahead please!")
	t.Love(commandExecutedProperly)
	return
}
