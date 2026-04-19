package sender

import (
	"context"
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
	return func(_ context.Context, transaction *msmtpd.Transaction) error {
		domain := strings.Split(transaction.MailFrom.Address, "@")[1]
		_, found := goodDomains[domain]
		if found {
			transaction.LogInfo("Sender's %s domain is whitelisted", transaction.MailFrom.String())
			return nil
		}
		_, found = goodMailFroms[transaction.MailFrom]
		if found {
			transaction.LogInfo("Sender %s is whitelisted", transaction.MailFrom.String())
			return nil
		}
		transaction.LogInfo("Sender %s is not whitelisted", transaction.MailFrom.String())
		return msmtpd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but your email address is not in whitelist",
		}
	}
}

// AcceptMailFromDomains allows all senders from domain list provided
func AcceptMailFromDomains(whitelist []string) msmtpd.SenderChecker {
	return AcceptMailFromDomainsOrAddresses(whitelist, nil)
}

// AcceptMailFromAddresses is filter to accept emails only from predefined whilelist of addresses,
// it is more strict version of AcceptMailFromDomains which can accept email from every mailbox of domain
func AcceptMailFromAddresses(whitelist []string) msmtpd.SenderChecker {
	return AcceptMailFromDomainsOrAddresses(nil, whitelist)
}
