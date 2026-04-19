package helo

import (
	"context"

	"github.com/vodolaz095/msmtpd"
)

// HateForRDNSMismatch is less strict version of DenyReverseDNSMismatch that only applies negative karma for failed RDNS check
func HateForRDNSMismatch(howMuch uint) msmtpd.HelloChecker {
	return func(ctx context.Context, transaction *msmtpd.Transaction) (err error) {
		err = DenyReverseDNSMismatch(ctx, transaction)
		if err == complain {
			newHateLevel := transaction.Hate(int(howMuch))
			transaction.LogInfo("giving %v hate for RDNS mismatch, new level is %v",
				howMuch, newHateLevel)
			return nil
		}
		return err
	}
}
