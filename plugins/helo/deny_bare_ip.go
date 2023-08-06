package helo

import (
	"net"

	"github.com/vodolaz095/msmtpd"
)

// DenyBareIP denies clients which provide bare IP address in HELO/EHLO command
func DenyBareIP(transaction *msmtpd.Transaction) error {
	if transaction.IsFlagSet(IsLocalAddressFlagName) {
		transaction.LogDebug("Connecting from local address %s, DenyBareIP check disabled",
			transaction.Addr.String())
		return nil
	}

	if net.ParseIP(transaction.HeloName) != nil {
		transaction.LogWarn("HELO/EHLO hostname %s is bare ip", transaction.HeloName)
		return complain
	}
	transaction.LogDebug("HELO/EHLO %s seems to be not bare IP", transaction.HeloName)
	return nil
}
