package connection

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/vodolaz095/msmtpd"
)

// CheckByReverseIPBlacklists checks Transaction IP address against Reverse IP Blacklists provided. If IP address
// presents in lists more times than tolerance, connection is blocked
func CheckByReverseIPBlacklists(tolerance uint32, lists []string) msmtpd.ConnectionChecker {
	return func(tr *msmtpd.Transaction) error {
		var listed uint32
		ip := tr.Addr.(*net.TCPAddr).IP
		reversed, err := reverse(ip)
		if err != nil {
			tr.LogError(err, fmt.Sprintf("while reversing transaction address %s", tr.Addr.String()))
			return msmtpd.ErrServiceNotAvailable
		}
		wg := sync.WaitGroup{}
		wg.Add(len(lists))
		for j := range lists {
			go func(t *msmtpd.Transaction, rr, list string) {
				defer wg.Done()
				hostname := fmt.Sprintf("%s.%s", rr, list)
				t.LogDebug("Checking %s...", hostname)
				names, errR := t.Resolver().LookupHost(t.Context(), hostname)
				if errR != nil {
					if !strings.Contains(errR.Error(), "no such host") {
						t.LogError(errR, "while resolving "+hostname)
					}
					return
				}
				if len(names) == 0 {
					return
				}
				switch names[0] {
				case "127.0.0.2", "127.0.0.3", "127.0.0.4":
					tr.LogDebug("%s is listed in %s", ip, list)
					atomic.AddUint32(&listed, 1)
					break
				default:
					tr.LogDebug("%s is not listed in %s", ip, list)
				}
			}(tr, reversed, lists[j])
		}
		wg.Wait()
		if listed > tolerance {
			tr.LogWarn("Address %s is listed in %v reverse ip blacklists of %v provided",
				ip, listed, len(lists),
			)
			return msmtpd.ErrServiceNotAvailable
		}
		tr.LogInfo("Address %s is not listed in %v blacklists provided",
			ip, len(lists),
		)
		return nil
	}
}
