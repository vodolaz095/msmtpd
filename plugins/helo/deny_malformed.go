package helo

import (
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// DenyMalformedDomain checks, if domain in HELO request belongs to top list domains like .ru, .su and so on
func DenyMalformedDomain(transaction *msmtpd.Transaction) error {
	var pass bool
	if transaction.IsFlagSet(IsLocalAddressFlagName) {
		transaction.LogDebug("Connecting from local address %s, DenyMalformedDomain check disabled",
			transaction.Addr.String())
		return nil
	}
	fixed := strings.ToUpper(transaction.HeloName)
	for i := range TopListDomains {
		if pass {
			continue
		}
		if strings.HasSuffix(fixed, "."+TopListDomains[i]) {
			pass = true
		}
		if strings.HasSuffix(fixed, "."+TopListDomains[i]+".") {
			pass = true
		}
	}
	if !pass {
		transaction.LogWarn("HELO/EHLO hostname %s is invalid", transaction.HeloName)
		return complain
	}
	transaction.LogDebug("HELO/EHLO %s seems to be in top domain list", transaction.HeloName)
	return nil
}
