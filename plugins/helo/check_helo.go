package helo

import (
	"net"
	"net/netip"
	"strings"

	"msmtpd"
)

// Options used to tune how we check HELO/EHLO
type Options struct {
	// TolerateInvalidHostname allows HELO/EHLO argument to be invalid hostname
	TolerateInvalidHostname bool
	// TolerateBareIP allows HELO/EHLO argument to be bare IP
	TolerateBareIP bool
	// TolerateDynamic allows HELO/EHLO argument to include parts of connection IP address
	TolerateDynamic bool
	// TolerateRDNSMismatch allows reverse DNS (PTR) server name of connecting IP to be different, then HELO provided
	TolerateRDNSMismatch bool
	// IgnoreHostnameForLocalAddresses allows to provide wierd hostnames in HELO/EHLO from local ip ranges
	IgnoreHostnameForLocalAddresses bool
}

const complain = "I don't like the way you introduce yourself. Goodbye!"

// CheckHELO is standard msmtpd.HelloChecker to check how client introduce itself
func CheckHELO(opts Options) msmtpd.HelloChecker {
	tlds := strings.Split(topListDomains, "\n")
	return func(transaction *msmtpd.Transaction) error {
		var pass bool
		if opts.IgnoreHostnameForLocalAddresses {
			addrPort, err := netip.ParseAddrPort(transaction.Addr.String())
			if err != nil {
				transaction.LogError(err, "while parsing remote address "+transaction.Addr.String())
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			if addrPort.Addr().IsPrivate() {
				transaction.LogDebug("Since clients address %s is local, HELO/EHLO %s will work",
					transaction.Addr.String(), transaction.HeloName,
				)
				return nil
			}
		}
		if !opts.TolerateInvalidHostname {
			fixed := strings.ToUpper(transaction.HeloName)
			for i := range tlds {
				if pass {
					continue
				}
				if strings.HasSuffix(fixed, "."+tlds[i]) {
					pass = true
				}
			}
			if !pass {
				transaction.LogWarn("HELO/EHLO hostname %s is invalid", transaction.HeloName)
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			transaction.LogDebug("HELO/EHLO %s seems to be in top domain list", transaction.HeloName)
		}
		if !opts.TolerateBareIP {
			if net.ParseIP(transaction.HeloName) != nil {
				transaction.LogWarn("HELO/EHLO hostname %s is bare ip", transaction.HeloName)
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			transaction.LogDebug("HELO/EHLO %s seems to be not bare IP", transaction.HeloName)
		}
		if !opts.TolerateDynamic {
			if isDynamic(transaction) {
				transaction.LogWarn("HELO/EHLO hostname %s looks dynamic", transaction.HeloName)
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			transaction.LogDebug("HELO/EHLO hostname %s looks static", transaction.HeloName)
		}
		if !opts.TolerateRDNSMismatch {
			pass = false
			needle := transaction.Addr.(*net.TCPAddr).IP.String()
			transaction.LogDebug("Resolving PTR for %s...", needle)
			for i := range transaction.PTRs {
				transaction.LogDebug("For %s PTR %s is resolved", needle, transaction.PTRs[i])
				if transaction.PTRs[i] == transaction.HeloName+"." {
					pass = true
				}
			}
			if !pass {
				transaction.LogWarn("For HELO/EHLO %s there is no matching PTR records", transaction.HeloName)
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			transaction.LogInfo("HELO/EHLO %s is matching RDNS record", transaction.HeloName)
		}
		return nil
	}
}
