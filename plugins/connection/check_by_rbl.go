package connection

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/vodolaz095/msmtpd"
)

// CheckByReverseIPBlacklists checks Transaction IP address against Reverse IP Blacklists provided. If IP address
// presents in lists more times than tolerance, connection is blocked
func CheckByReverseIPBlacklists(tolerance uint32, lists []string) msmtpd.ConnectionChecker {
	return func(ctx context.Context, tr *msmtpd.Transaction) error {
		var listed uint32
		ip := tr.Addr.(*net.TCPAddr).IP
		reversed, err := reverse(ip)
		if err != nil {
			tr.LogError(err, fmt.Sprintf("while reversing transaction address %s", tr.Addr.String()))
			return msmtpd.ErrServiceNotAvailable
		}
		eg, ctx2 := errgroup.WithContext(ctx)
		for j := range lists {
			eg.Go(func() error {
				hostname := fmt.Sprintf("%s.%s", reversed, lists[j])
				tr.LogDebug("Checking %s...", hostname)
				names, errR := tr.Resolver().LookupHost(ctx2, hostname)
				if errR != nil {
					if !strings.Contains(errR.Error(), "no such host") {
						tr.LogError(errR, "while resolving "+hostname)
					}
					return errR
				}
				if len(names) == 0 {
					return nil
				}
				switch names[0] {
				case "127.0.0.2", "127.0.0.3", "127.0.0.4":
					tr.LogDebug("%s is listed in %s", ip, lists[j])
					atomic.AddUint32(&listed, 1)
					break
				default:
					tr.LogDebug("%s is not listed in %s", ip, lists[j])
				}
				return nil
			})
		}
		err = eg.Wait()
		if err != nil {
			return msmtpd.ErrServiceNotAvailable
		}

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
