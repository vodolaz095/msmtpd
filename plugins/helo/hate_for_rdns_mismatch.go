package helo

import "github.com/vodolaz095/msmtpd"

// HateForRDNSMismatch is less strict version of DenyReverseDNSMismatch that only applies negative karma for failed RDNS check
func HateForRDNSMismatch(howMuch uint) msmtpd.HelloChecker {
	return func(transaction *msmtpd.Transaction) (err error) {
		err = DenyReverseDNSMismatch(transaction)
		if err == complain {
			newHateLevel := transaction.Hate(int(howMuch))
			transaction.LogInfo("giving %v hate for RDNS mismatch, new level is %v",
				howMuch, newHateLevel)
			return nil
		}
		return err
	}
}
