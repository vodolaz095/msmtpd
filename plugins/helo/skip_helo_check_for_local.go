package helo

import (
	"context"
	"net/netip"

	"github.com/vodolaz095/msmtpd"
)

// SkipHeloCheckForLocal allows local clients provide anything in HELO/EHLO
func SkipHeloCheckForLocal(_ context.Context, transaction *msmtpd.Transaction) error {
	addrPort, err := netip.ParseAddrPort(transaction.Addr.String())
	if err != nil {
		transaction.LogError(err, "while parsing remote address "+transaction.Addr.String())
		return complain
	}
	if addrPort.Addr().IsLoopback() {
		transaction.LogInfo("Skipping HELO/EHLO checks for loopback address %s and HELO %s",
			transaction.Addr.String(), transaction.HeloName,
		)
		transaction.SetFlag(IsLocalAddressFlagName)
	}
	if addrPort.Addr().IsLinkLocalUnicast() {
		transaction.LogInfo("Skipping HELO/EHLO checks for local unicast address %s and HELO %s",
			transaction.Addr.String(), transaction.HeloName,
		)
		transaction.SetFlag(IsLocalAddressFlagName)
	}
	if addrPort.Addr().IsPrivate() {
		transaction.LogInfo("Skipping HELO/EHLO checks for private network address %s and HELO %s",
			transaction.Addr.String(), transaction.HeloName,
		)
		transaction.SetFlag(IsLocalAddressFlagName)
	}
	return nil
}
