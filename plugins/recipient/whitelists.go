package recipient

import (
	"log"
	"net/mail"
	"strings"

	"msmtpd"
)

// AcceptMailForDomainsOrAddresses is msmtpd.RecipientChecker function that accepts emails either for anything on domain list, or to predefined list of email addresses
func AcceptMailForDomainsOrAddresses(whitelistedDomains, whitelistedAddresses []string) msmtpd.RecipientChecker {
	var err error
	var parsed *mail.Address
	goodRecipients := make(map[mail.Address]bool, 0)

	for i, raw := range whitelistedAddresses {
		parsed, err = mail.ParseAddress(raw)
		if err != nil {
			log.Fatalf("%s : while plugin rcpt_to/AcceptMailForDomainsOrAddresses tries to parse address %v %s",
				err, i, raw,
			)
		}
		goodRecipients[*parsed] = true
	}
	goodDomains := make(map[string]bool, 0)
	for _, raw := range whitelistedDomains {
		goodDomains[strings.ToLower(raw)] = true
	}
	return func(transaction *msmtpd.Transaction, recipient *mail.Address) error {
		domain := strings.Split(recipient.Address, "@")[1]
		_, found := goodDomains[domain]
		if found {
			transaction.LogDebug("Recipient's %s domain is whitelisted", transaction.MailFrom.String())
			return nil
		}
		_, found = goodRecipients[*recipient]
		if found {
			transaction.LogDebug("Recipient %s is whitelisted", transaction.MailFrom.String())
			return nil
		}
		return msmtpd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but recipient's email address is not in whitelist",
		}
	}
}

// AcceptMailForDomains is msmtpd.RecipientChecker function that accepts emails either for anything on domain list
func AcceptMailForDomains(whitelist []string) msmtpd.RecipientChecker {
	return AcceptMailForDomainsOrAddresses(whitelist, nil)
}

// AcceptMailForAddresses is msmtpd.RecipientChecker function that accepts emails either for anything in list of addresses
func AcceptMailForAddresses(whitelist []string) msmtpd.RecipientChecker {
	return AcceptMailForDomainsOrAddresses(nil, whitelist)
}
