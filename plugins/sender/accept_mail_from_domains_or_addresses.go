package sender

import (
	"log"
	"net/mail"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// AcceptMailFromDomainsOrAddresses allows senders either from one of whilelisted domain, or from one of whitelisted addresses.
// It is more complicated version of AcceptMailFromDomains and AcceptMailFromAddresses.
func AcceptMailFromDomainsOrAddresses(whitelistedDomains, whitelistedAddresses []string) msmtpd.SenderChecker {
	var err error
	var parsed *mail.Address
	goodMailFroms := make(map[mail.Address]bool, 0)

	for i, raw := range whitelistedAddresses {
		parsed, err = mail.ParseAddress(raw)
		if err != nil {
			log.Fatalf("%s : while plugin mail_from/AcceptMailFromDomainsOrAddresses tries to parse address %v %s",
				err, i, raw,
			)
		}
		goodMailFroms[*parsed] = true
	}
	goodDomains := make(map[string]bool, 0)
	for _, raw := range whitelistedDomains {
		goodDomains[strings.ToLower(raw)] = true
	}
	return func(transaction *msmtpd.Transaction) error {
		domain := strings.Split(transaction.MailFrom.Address, "@")[1]
		_, found := goodDomains[domain]
		if found {
			transaction.LogDebug("Sender's %s domain is whitelisted", transaction.MailFrom.String())
			return nil
		}
		_, found = goodMailFroms[transaction.MailFrom]
		if found {
			transaction.LogDebug("Sender %s is whitelisted", transaction.MailFrom.String())
			return nil
		}
		return msmtpd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but your email address is not in whitelist",
		}
	}
}
