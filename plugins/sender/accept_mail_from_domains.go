package sender

import (
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// AcceptMailFromDomains allows all senders from domain list provided
func AcceptMailFromDomains(whitelist []string) msmtpd.SenderChecker {
	goodDomains := make(map[string]bool, 0)
	for _, raw := range whitelist {
		goodDomains[strings.ToLower(raw)] = true
	}
	return func(transaction *msmtpd.Transaction) error {
		domain := strings.Split(transaction.MailFrom.Address, "@")[1]
		_, found := goodDomains[domain]
		if found {
			transaction.LogDebug("Sender's %s domain is whitelisted", transaction.MailFrom.String())
			return nil
		}
		return msmtpd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but your email address is not in whitelist",
		}
	}
}
