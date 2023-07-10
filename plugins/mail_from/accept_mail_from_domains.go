package mail_from

import (
	"net/mail"
	"strings"

	"msmtpd"
)

// AcceptMailFromDomains allows all senders from domain list provided
func AcceptMailFromDomains(whitelist []string) msmptd.CheckerFunc {
	goodDomains := make(map[string]bool, 0)
	for _, raw := range whitelist {
		goodDomains[strings.ToLower(raw)] = true
	}
	return func(transaction *msmptd.Transaction, name string) error {
		addr, err := mail.ParseAddress(name)
		if err != nil {
			transaction.LogWarn("%s : while parsing %s as email address", err, name)
			return msmptd.ErrorSMTP{Code: 502, Message: "Malformed e-mail address"}
		}
		domain := strings.Split(addr.Address, "@")[1]
		_, found := goodDomains[domain]
		if found {
			transaction.LogDebug("Sender's %s domain is whitelisted", addr.String())
			return nil
		}
		return msmptd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but your email address is not in whitelist",
		}
	}
}
