package mail_from

import (
	"strings"

	"msmtpd"
)

func AcceptMailFromDomains(whitelist []string) msmptd.CheckerFunc {
	goodDomains := make(map[string]bool, 0)
	for _, raw := range whitelist {
		goodDomains[raw] = true
	}

	return func(transaction *msmptd.Transaction, name string) error {
		// check if we already whitelisted this address
		status, found := transaction.GetFact(MailFromWhileListed)
		if found {
			if status == "true" {
				return nil
			}
		}

		domain := strings.Split(transaction.MailFrom.Address, "@")[1]
		_, found = goodDomains[domain]
		if found {
			transaction.LogDebug("Sender's %s domain is whitelisted", transaction.MailFrom.String())
			transaction.SetFact(MailFromWhileListed, "true")
			return nil
		}
		return msmptd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but your email address is not in whitelist",
		}
	}
}
