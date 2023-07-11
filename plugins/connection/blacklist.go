package connection

import (
	"msmtpd"
)

func Whitelist(ipAddressesToAccept []string) func(transaction *msmtpd.Transaction) error {
	return func(transaction *msmtpd.Transaction) error {
		return nil
	}
}

func Blacklist(ipAddressesToBlock []string) func(transaction *msmtpd.Transaction) error {
	return func(transaction *msmtpd.Transaction) error {
		return nil
	}
}
