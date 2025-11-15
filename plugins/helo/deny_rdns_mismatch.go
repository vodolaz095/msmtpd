package helo

import (
	"context"
	"fmt"
	"net"

	"github.com/vodolaz095/msmtpd"
)

// DenyReverseDNSMismatch is complicated and very strict test which ensures that
// 1. HELO/EHLO matches any of PTR records resolved for connecting IP
// 2. PTR records of connecting IP are resolved (aka have DNS A records) into IP addresses including connecting IP
// This test prevents delivery from majority small GI domains, which are known to be spammy.
func DenyReverseDNSMismatch(ctx context.Context, transaction *msmtpd.Transaction) (err error) {
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
	// now we resolve IP addresses of PTR records
	found = false
	var resolvedPtrs []string
	for i := range transaction.PTRs {
		addrs, errLookup := transaction.Resolver().
			LookupHost(ctx, transaction.PTRs[i])
		if errLookup != nil {
			transaction.LogError(err, fmt.Sprintf("while resolving A record for PTR of %s", transaction.PTRs[i]))
			return complain
		}
		resolvedPtrs = append(resolvedPtrs, addrs...)
		transaction.LogDebug("For PTR %s this addresses resolved %v",
			transaction.PTRs[i], addrs,
		)
	}
	raw := transaction.Addr.(*net.TCPAddr).IP
	for i := range resolvedPtrs {
		if resolvedPtrs[i] == raw.String() {
			found = true
			transaction.LogDebug("PTR %s resolves into connecting address %s",
				resolvedPtrs[i], raw.String(),
			)
		}
	}
	if !found {
		transaction.LogInfo("DNS-PTR mismatch! PTR records %v resolves into %v - remote IP %s is not among them",
			transaction.PTRs, resolvedPtrs, raw.String(),
		)
		return complain
	}
	transaction.LogInfo("DNS-PTR passes PTR records %v resolves into %v - remote IP %s included!",
		transaction.PTRs, resolvedPtrs, raw.String(),
	)
	return nil
}
