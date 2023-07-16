package mail_from

import (
	"net"
	"strings"

	"msmtpd"
)

// Good read - https://en.wikipedia.org/wiki/MX_record

// SenderIsResolvableOptions are options for SenderIsResolvable plugin, default values are recommended
type SenderIsResolvableOptions struct {
	// FallbackToAddressRecord allows to delivery to 25th port of address resolved by A or AAAA
	// queries if MX records are not present
	// See https://en.wikipedia.org/wiki/MX_record#Fallback_to_the_address_record
	FallbackToAddressRecord bool

	// AllowLocalAddresses means we accept email from local IP addresses
	// according to https://rfc-editor.org/rfc/rfc1918.html (IPv4 addresses) and
	// https://rfc-editor.org/rfc/rfc4193.html (IPv6 addresses).
	AllowLocalAddresses bool
}

const SenderIsNotResolvableComplain = "Seems like i cannot find your sender address mail servers using DNS, please, try again later"

func SenderIsResolvable(opts SenderIsResolvableOptions) msmtpd.SenderChecker {
	return func(transaction *msmtpd.Transaction) error {
		possibleMxServers := make([]string, 0)
		usableMxServers := make([]net.IP, 0)
		resolver := transaction.Resolver()
		ctx := transaction.Context()
		domain := strings.Split(transaction.MailFrom.Address, "@")[1]
		mxRecords, err := resolver.LookupMX(ctx, domain)
		if err != nil {
			transaction.LogWarn("%s : while resolving MX records for domain %s of %s",
				err, domain, transaction.MailFrom.String())
			mxRecords = nil
		}
		if len(mxRecords) > 0 {
			for _, record := range mxRecords {
				transaction.LogDebug("For domain %s MX record %s %v is found",
					domain, record.Host, record.Pref)
				possibleMxServers = append(possibleMxServers, record.Host)
			}
		} else {
			if opts.FallbackToAddressRecord {
				transaction.LogDebug("For domain %s no MX records found, using A/AAAA record fallback",
					domain)
				possibleMxServers = append(possibleMxServers, domain)
			} else {
				return msmtpd.ErrorSMTP{
					Code:    421,
					Message: SenderIsNotResolvableComplain,
				}
			}
		}
		if len(possibleMxServers) == 0 {
			transaction.LogDebug("For domain %s there are no possible email exchanges", domain)
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: SenderIsNotResolvableComplain,
			}
		}
		transaction.LogDebug("For domain %s there are %v possible email exchanges",
			domain, len(possibleMxServers))

		for _, record := range possibleMxServers {
			transaction.LogDebug("Checking mx server %s of domain %s", record, domain)
			ips, errLookUp := resolver.LookupIP(ctx, "ip", record)
			if errLookUp != nil {
				transaction.LogWarn("%s : while resolving IP address for mailserver %s of domain of %s for %s",
					errLookUp, record, domain, transaction.MailFrom.String())
				continue
			}
			for _, ip := range ips {
				transaction.LogDebug("Checking mx server %s ip %s of domain %s...",
					record, ip.String(), domain)
				if ip.IsPrivate() {
					if opts.AllowLocalAddresses {
						transaction.LogDebug("IP address %s of DNS server %s of domain %s seems legit, even if it is local",
							ip.String(), record, domain,
						)
						usableMxServers = append(usableMxServers, ip)
						continue
					}
				}
				if ip.IsLoopback() {
					transaction.LogDebug("MX server %s ip %s of domain %s is loopback!",
						record, ip.String(), domain)
					continue
				}
				if ip.IsUnspecified() {
					transaction.LogDebug("MX server %s ip %s of domain %s is uspecified!",
						record, ip.String(), domain)
					continue
				}
				if ip.IsMulticast() {
					transaction.LogDebug("MX server %s ip %s of domain %s is Multicast!",
						record, ip.String(), domain)
					continue
				}
				if ip.IsGlobalUnicast() {
					transaction.LogDebug("MX server %s ip %s of domain %s is Global Unicast!",
						record, ip.String(), domain)
					usableMxServers = append(usableMxServers, ip)
					continue
				}
				if ip.IsLinkLocalUnicast() {
					transaction.LogDebug("MX server %s ip %s of domain %s is Link Local Unicast!",
						record, ip.String(), domain)
					continue
				}
				if ip.IsInterfaceLocalMulticast() {
					transaction.LogDebug("MX server %s ip %s of domain %s is Interface Local Multicast!",
						record, ip.String(), domain)
					continue
				}
				transaction.LogDebug("IP address %s of DNS server %s of domain %s seems legit",
					ip.String(), record, domain,
				)
				usableMxServers = append(usableMxServers, ip)
			}
		}
		if len(usableMxServers) > 0 {
			transaction.LogDebug("We found %v usable MX servers for domain %s",
				len(usableMxServers), domain)
			return nil
		}
		return msmtpd.ErrorSMTP{
			Code:    421,
			Message: SenderIsNotResolvableComplain,
		}
	}
}
