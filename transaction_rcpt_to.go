package msmtpd

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

func (t *Transaction) handleRCPT(cmd command) {
	if len(cmd.params) != 2 || strings.ToUpper(cmd.params[0]) != "TO" {
		t.Hate(missingParameterPenalty)
		t.reply(502, "Invalid syntax.")
		return
	}
	if t.HeloName == "" {
		t.LogDebug("RCPT TO called without HELO/EHLO")
		t.Hate(missingParameterPenalty)
		t.reply(502, "Please introduce yourself first.")
		return
	}
	if !t.Encrypted && t.server.ForceTLS {
		t.LogDebug("RCPT TO called without STARTTLS")
		t.Hate(missingParameterPenalty)
		t.reply(502, "Please turn on TLS by issuing a STARTTLS command.")
		return
	}
	if t.server.Authenticator != nil && t.Username == "" {
		t.LogDebug("RCPT TO called without authentication")
		t.Hate(missingParameterPenalty)
		t.reply(530, "Authentication Required.")
		return
	}
	if t.MailFrom.Address == "" {
		t.LogDebug("RCPT TO called without MAIL FROM")
		t.Hate(missingParameterPenalty)
		t.reply(502, "It seems you haven't called MAIL FROM in order to explain who sends your message.")
		return
	}
	if len(t.RcptTo) >= t.server.MaxRecipients {
		t.LogDebug("Too many recipients")
		t.Hate(tooManyRecipientsPenalty)
		t.reply(452, "Too many recipients")
		return
	}
	addr, err := parseAddress(cmd.params[1])
	if err != nil {
		t.Hate(missingParameterPenalty)
		t.reply(502, "Malformed e-mail address")
		return
	}
	t.LogDebug("Checking recipient %s by %v RecipientCheckers...",
		addr.String(), len(t.server.RecipientCheckers))
	for k := range t.server.RecipientCheckers {
		err = t.server.RecipientCheckers[k](t, addr)
		if err != nil {
			t.Hate(unknownRecipientPenalty)
			t.error(err)
			return
		}
	}
	t.RcptTo = append(t.RcptTo, *addr)
	switch len(t.RcptTo) {
	case 1:
		t.LogInfo("Recipient %s will be 1st one in transaction", addr)
		break
	case 2:
		t.LogInfo("Recipient %s will be 2nd one in transaction", addr)
		break
	case 3:
		t.LogInfo("Recipient %s will be 2nd one in transaction", addr)
		break
	default:
		t.LogInfo("Recipient %s will be %dth one in transaction", addr, len(t.RcptTo))
	}
	recipientsAsStrings := make([]string, len(t.RcptTo))
	for i := range t.RcptTo {
		recipientsAsStrings[i] = t.RcptTo[i].String()
	}
	t.Span.SetAttributes(attribute.StringSlice("to", recipientsAsStrings))
	aliasesAsStrings := make([]string, len(t.Aliases))
	for i := range t.Aliases {
		aliasesAsStrings[i] = t.Aliases[i].String()
	}
	t.Span.SetAttributes(attribute.StringSlice("aliases", aliasesAsStrings))
	t.reply(250, "It seems i can handle delivery for this recipient, i'll do my best!")
	if len(t.RcptTo) == 1 { // too many recipients should not give too many love for transaction
		t.Love(commandExecutedProperly)
	}
	return
}
