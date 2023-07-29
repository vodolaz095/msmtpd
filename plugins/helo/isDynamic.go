package helo

import (
	"fmt"
	"net"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// isDynamic returns true, if Transaction.Addr seems like dynamic IP address
// TODO: i cannot say i checked this function on all possible cases
func isDynamic(tr *msmtpd.Transaction) (yes bool) {
	var needles []string
	raw := tr.Addr.(*net.TCPAddr).IP
	if raw.To4() == nil {
		// i haven't encountered ham being send from IPv6
		tr.LogDebug("IP %s looks like IPv6", raw.String())
		return true
	}
	octets := strings.Split(raw.To4().String(), ".")
	if len(octets) != 4 {
		tr.LogDebug("IP %s has no IPv4 octets", raw.String())
		return true
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
		tr.LogTrace("Checking `%s` to cointain `%s`", tr.HeloName, needles[i])
		if strings.Contains(tr.HeloName, needles[i]) {
			tr.LogDebug("IP address %s triggered needle %v %s",
				raw.String(), i, needles[i])
			yes = true
			break
		}
	}
	return yes
}
