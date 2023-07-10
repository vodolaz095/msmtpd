package msmtpd

import "strings"

func (t *Transaction) handleRCPT(cmd command) {
	if len(cmd.params) != 2 || strings.ToUpper(cmd.params[0]) != "TO" {
		t.reply(502, "Invalid syntax.")
		return
	}
	if t.MailFrom.Address == "" {
		t.reply(502, "It seems you haven't called MAIL FROM in order to explain who sends your message.")
		return
	}
	if len(t.RcptTo) >= t.server.MaxRecipients {
		t.reply(452, "Too many recipients")
		return
	}
	addr, err := parseAddress(cmd.params[1])
	if err != nil {
		t.reply(502, "Malformed e-mail address")
		return
	}
	t.LogDebug("Checking recipient %s...", addr)
	for k := range t.server.RecipientCheckers {
		err = t.server.RecipientCheckers[k](t, addr)
		if err != nil {
			t.error(err)
			return
		}
	}
	t.RcptTo = append(t.RcptTo, *addr)
	t.LogInfo("Recipient %s will be %v one in transaction", addr, len(t.RcptTo))
	t.reply(250, "It seems i can handle delivery for this recipient, i'll do my best!")
	return
}
