package connection

import (
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// DenyPTRs allows to deny senders if their PTR records has suffix from list, for example,
// we can restrict all Amazon Simple Email Service by providing suffix "amazonses.com.",
func DenyPTRs(listOfPtrSuffixes []string) msmtpd.ConnectionChecker {
	return func(tr *msmtpd.Transaction) error {
		var i, j int
		var bad bool
		for i = range tr.PTRs {
			if bad {
				break
			}
			for j = range listOfPtrSuffixes {
				if strings.HasSuffix(tr.PTRs[i], listOfPtrSuffixes[j]) {
					tr.LogInfo("PTR %v %s has bad suffix %v %s",
						i, tr.PTRs[i], j, listOfPtrSuffixes[j],
					)
					bad = true
					break
				}
			}
		}
		if bad {
			return msmtpd.ErrorSMTP{
				Code:    521,
				Message: "Your IP address is blacklisted. Sorry. You can cry me a river.", // lol
			}
		}
		tr.LogInfo("PTRs %v looks ok", tr.PTRs)
		return nil
	}
}
