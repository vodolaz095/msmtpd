package helo

import "github.com/vodolaz095/msmtpd"

func DenyReverseDNSMismatch(transaction *msmtpd.Transaction) (err error) {
	var found bool
	if transaction.IsFlagSet(IsLocalAddressFlagName) {
		transaction.LogDebug("Connecting from local address %s, DenyReverseDNSMismatch check disabled",
			transaction.Addr.String())
		return nil
	}
	if len(transaction.PTRs) == 0 {
		transaction.LogWarn("Address %s has no PTR records - dns mismatch detected", transaction.Addr.String())
		return complain
	}
	for i := range transaction.PTRs {
		if transaction.PTRs[i] == transaction.HeloName {
			transaction.LogInfo("HELO/EHLO %s is matching RDNS record %s",
				transaction.HeloName, transaction.PTRs[i])
			found = true
			break
		}
		if transaction.PTRs[i] == transaction.HeloName+"." {
			transaction.LogInfo("HELO/EHLO %s. is matching RDNS record %s",
				transaction.HeloName, transaction.PTRs[i])
			found = true
			break
		}
	}
	if !found {
		transaction.LogWarn("For HELO/EHLO %s there is no matching PTR records", transaction.HeloName)
		return complain
	}
	return nil
}
