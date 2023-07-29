package connection

import (
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// Whitelist prevents connecting from remote addresses in not present in list
func Whitelist(ipAddressesToAccept []string) func(transaction *msmtpd.Transaction) error {
	return func(transaction *msmtpd.Transaction) error {
		var found bool
		for i := range ipAddressesToAccept {
			found = strings.HasPrefix(transaction.Addr.String(), ipAddressesToAccept[i])
			if found {
				transaction.LogInfo("IP address %s is whitelisted by rule %s",
					transaction.Addr.String(), ipAddressesToAccept[i],
				)
				break
			}
		}
		if found {
			return nil
		}
		transaction.LogDebug("IP address %s is not whitelisted", transaction.Addr.String())
		return msmtpd.ErrorSMTP{
			Code:    521,
			Message: "FUCK OFF!", // lol
		}
	}
}

// Blacklist prevents connecting from remote addresses in list
func Blacklist(ipAddressesToBlock []string) func(transaction *msmtpd.Transaction) error {
	return func(transaction *msmtpd.Transaction) error {
		var found bool
		for i := range ipAddressesToBlock {
			found = strings.HasPrefix(transaction.Addr.String(), ipAddressesToBlock[i])
			if found {
				transaction.LogInfo("IP address %s is blacklisted by rule %s",
					transaction.Addr.String(), ipAddressesToBlock[i],
				)
				break
			}
		}
		if found {
			return msmtpd.ErrorSMTP{
				Code:    521,
				Message: "FUCK OFF!", // lol
			}
		}
		transaction.LogDebug("IP address %s is not blacklisted", transaction.Addr.String())
		return nil
	}
}
