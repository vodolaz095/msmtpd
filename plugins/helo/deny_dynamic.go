package helo

import (
	"fmt"
	"net"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

func DenyDynamicIP(transaction *msmtpd.Transaction) error {
	if transaction.IsFlagSet(IsLocalAddressFlagName) {
		transaction.LogDebug("Connecting from local address %s, DenyBareIP check disabled",
			transaction.Addr.String())
		return nil
	}
	var isDynamic bool
	var needles []string
	raw := transaction.Addr.(*net.TCPAddr).IP
	if raw.To4() == nil {
		// i haven't encountered ham being send from IPv6
		transaction.LogDebug("IP %s looks like IPv6", raw.String())
		return complain
	}
	octets := strings.Split(raw.To4().String(), ".")
	if len(octets) != 4 {
		transaction.LogDebug("IP %s has no IPv4 octets", raw.String())
		return complain
	}
	needles = append(needles, fmt.Sprintf("%s-%s-%s-%s", octets[0], octets[1], octets[2], octets[3]))
	needles = append(needles, fmt.Sprintf("%s-%s-%s-%s", octets[3], octets[2], octets[1], octets[0]))
	needles = append(needles, fmt.Sprintf("%s-%s-%s", octets[0], octets[1], octets[2]))
	needles = append(needles, fmt.Sprintf("%s-%s-%s", octets[2], octets[1], octets[0]))
	needles = append(needles, fmt.Sprintf("%s-%s", octets[0], octets[1]))
	needles = append(needles, fmt.Sprintf("%s-%s", octets[1], octets[0]))

	needles = append(needles, fmt.Sprintf("%s.%s.%s.%s", octets[0], octets[1], octets[2], octets[3]))
	needles = append(needles, fmt.Sprintf("%s.%s.%s.%s", octets[3], octets[2], octets[1], octets[0]))
	needles = append(needles, fmt.Sprintf("%s.%s.%s", octets[0], octets[1], octets[2]))
	needles = append(needles, fmt.Sprintf("%s.%s.%s", octets[2], octets[1], octets[0]))
	needles = append(needles, fmt.Sprintf("%s.%s", octets[0], octets[1]))
	needles = append(needles, fmt.Sprintf("%s.%s", octets[1], octets[0]))
	for i := range needles {
		transaction.LogTrace("Checking `%s` to cointain `%s`", transaction.HeloName, needles[i])
		if strings.Contains(transaction.HeloName, needles[i]) {
			transaction.LogDebug("IP address %s triggered needle %v %s",
				raw.String(), i, needles[i])
			isDynamic = true
			break
		}
	}
	if isDynamic {
		transaction.LogWarn("HELO %s looks dynamic for address %s",
			transaction.HeloName, raw.To4().String())
		return complain
	}
	return nil
}
