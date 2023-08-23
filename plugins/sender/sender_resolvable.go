package sender

import (
	"net"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// Good read - https://en.wikipedia.org/wiki/MX_record

// IsResolvableOptions are options for IsResolvable plugin, default values are recommended
type IsResolvableOptions struct {
	// FallbackToAddressRecord allows to delivery to 25th port of address resolved by A or AAAA
	// queries if MX records are not present
	// See https://en.wikipedia.org/wiki/MX_record#Fallback_to_the_address_record
	FallbackToAddressRecord bool

	// AllowLocalAddresses means we accept email from domains, which MX records are resolved
	// into local IP addresses according to https://rfc-editor.org/rfc/rfc1918.html (IPv4 addresses) and
	// https://rfc-editor.org/rfc/rfc4193.html (IPv6 addresses).
	// For example, MX for something.example.org are `mx.something.example.org 10`, and
	// A record for mx.something.example.org is 192.168.1.3 - it is unusual case for
	// internet, but common for local networks.
	AllowLocalAddresses bool

	// AllowMxRecordToBeIP allows MX records to contain bare IP addresses, its unsafe, but still used
	AllowMxRecordToBeIP bool

	// DomainsToTrust is whitelist of domains you consider resolvable - for example, this is
	// local domain of your company with MX server having local IP in local network, and
	// on the same time, they are reachable from external network by white IPs.
	// So, IsResolvable should allow senders from this domains to be used for MAIL FROM
	DomainsToTrust []string
}

// IsNotResolvableComplain is human-readable thing we say to client with imaginary email address
const IsNotResolvableComplain = "Seems like i cannot find your sender address mail servers using DNS, please, try again later"

// IsResolvable is msmtpd.SenderChecker checker that performs DNS validations to proof we can send answer back to sender's email address
func IsResolvable(opts IsResolvableOptions) msmtpd.SenderChecker {
	return func(transaction *msmtpd.Transaction) error {
		domain := strings.Split(transaction.MailFrom.Address, "@")[1]

		var trustedDomain bool
		for i := range opts.DomainsToTrust {
			if domain == opts.DomainsToTrust[i] {
				trustedDomain = true
				break
			}
		}
		if trustedDomain {
			transaction.LogInfo("Sender %s is resolvable because he has trusted domain",
				transaction.MailFrom.Address,
			)
			return nil
		}

		possibleMxServers := make([]string, 0)     // A / AAAA records of possible MX servers
		availableMxServersIPs := make([]net.IP, 0) // IP addresses of possible MX servers
		usableMxServersIPs := make([]net.IP, 0)    // IP addresses of possible MX servers
		resolver := transaction.Resolver()
		ctx := transaction.Context()
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

				mxRecordAsIP := net.ParseIP(record.Host)
				if mxRecordAsIP == nil { // MX record is domain name
					possibleMxServers = append(possibleMxServers, record.Host)
				} else {
					// if MX record is raw IP address, and we allow it by config,
					// we add it to array for future checks
					if opts.AllowMxRecordToBeIP {
						availableMxServersIPs = append(availableMxServersIPs, mxRecordAsIP)
					}
				}
			}
		} else {
			if opts.FallbackToAddressRecord {
				transaction.LogDebug("For domain %s no MX records found, using A/AAAA record fallback",
					domain)
				possibleMxServers = append(possibleMxServers, domain)
			} else {
				return msmtpd.ErrorSMTP{
					Code:    421,
					Message: IsNotResolvableComplain,
				}
			}
		}
		if len(possibleMxServers) == 0 {
			transaction.LogInfo("For domain %s there are no possible email exchanges", domain)
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: IsNotResolvableComplain,
			}
		}
		transaction.LogDebug("For domain %s there are %v possible email exchanges",
			domain, len(possibleMxServers))

		for _, record := range possibleMxServers {
			transaction.LogDebug("Resolving IP of mx server %s of domain %s", record, domain)
			ips, errLookUp := resolver.LookupIP(ctx, "ip", record)
			if errLookUp != nil {
				transaction.LogWarn("%s : while resolving IP address for mailserver %s of domain of %s for %s",
					errLookUp, record, domain, transaction.MailFrom.String())
				continue
			}
			transaction.LogDebug("Checking mx server %s of domain %s having this IPs... %v",
				record, domain, ips)
			availableMxServersIPs = append(availableMxServersIPs, ips...)
		}

		transaction.LogDebug("For domain %s possible email exchanges %v were resolved into IP addresses: %v",
			domain, possibleMxServers, availableMxServersIPs)

		for _, ip := range availableMxServersIPs {
			if ip.IsPrivate() {
				if opts.AllowLocalAddresses {
					transaction.LogDebug("%s of MX of %s is local, but settings allows it",
						ip.String(), domain,
					)
					usableMxServersIPs = append(usableMxServersIPs, ip)
				} else {
					transaction.LogDebug("%s of MX of %s is local - unusual",
						ip.String(), domain,
					)
				}
				continue
			}
			if ip.IsLoopback() {
				transaction.LogDebug("%s of MX of %s is loopback - thanks for trolling :-)",
					ip.String(), domain,
				)
				continue
			}
			if ip.IsUnspecified() {
				transaction.LogDebug("%s of MX of %s is unspecified - useless",
					ip.String(), domain,
				)
				continue
			}
			if ip.IsMulticast() {
				transaction.LogDebug("%s of MX of %s is multicast - useless",
					ip.String(), domain,
				)
				continue
			}
			if ip.IsGlobalUnicast() {
				transaction.LogDebug("%s of MX of %s is global unicast - we can dial it",
					ip.String(), domain,
				)
				usableMxServersIPs = append(usableMxServersIPs, ip)
				continue
			}
			if ip.IsLinkLocalUnicast() {
				transaction.LogDebug("%s of MX of %s is link local unicast - weird",
					ip.String(), domain,
				)
				continue
			}
			if ip.IsInterfaceLocalMulticast() {
				transaction.LogDebug("%s of MX of %s is link local multicast - useless",
					ip.String(), domain,
				)
				continue
			}
			transaction.LogDebug("%s of MX of %s seems legit",
				ip.String(), domain,
			)
			usableMxServersIPs = append(usableMxServersIPs, ip)
		}

		if len(usableMxServersIPs) > 0 {
			transaction.LogInfo("We found %v usable MX servers for domain %s: %v",
				len(usableMxServersIPs), domain, usableMxServersIPs)
			return nil
		}
		transaction.LogInfo("No usable MX servers for domain %s", domain)
		return msmtpd.ErrorSMTP{
			Code:    421,
			Message: IsNotResolvableComplain,
		}
	}
}
