package connection

// Good read
// https://senderscore.org/mission/
// https://gist.github.com/agarzon/1a5a148ba0bade1033dc66716ebd98da
// https://socketloop.com/tutorials/golang-reverse-ip-address-for-reverse-dns-lookup-example

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

func reverse(ip net.IP) (reversed string, err error) {
	ip4 := ip.To4()
	if ip4 != nil {
		octets := strings.Split(ip4.String(), ".")
		reversed = fmt.Sprintf("%v.%v.%v.%v", octets[3], octets[2], octets[1], octets[0])
		return
	}
	err = fmt.Errorf("error parsing IPv4 %s", ip)
	return
}

const SenderscoreCounter = "senderscore"

// RequireSenderScore is connection checker which breaks connection if remote IP senderscore is too low.
// 0 - no info
// 0 -  70 You need to repair your email reputation. Your IP has been flagged for engaging in risky sending behaviors and your email performance could be suffering because of it.
// 70 - 80 You have a fine IP reputation score, but there’s room for improvement. Continue to follow industry best practices and optimize your email program.
// 80+  A history of healthy sending habits has resulted in a great email reputation. Good senders can get recognized and rewarded for their sending
func RequireSenderScore(minimalSenderScore uint) msmtpd.ConnectionChecker {
	if minimalSenderScore > 100 {
		panic("maximum sender score is 100")
	}

	return func(tr *msmtpd.Transaction) error {
		reversed, err := reverse(tr.Addr.(*net.TCPAddr).IP)
		if err != nil {
			tr.LogError(err, fmt.Sprintf("while reversing transaction address %s", tr.Addr.String()))
			return msmtpd.ErrServiceNotAvailable
		}
		domain := fmt.Sprintf("%s.score.senderscore.com", reversed)
		tr.LogDebug("Resolving %s...", domain)
		names, err := tr.Resolver().LookupHost(tr.Context(), domain)
		if err != nil {
			if strings.HasSuffix(err.Error(), "no such host") {
				tr.LogInfo("senderscore is 0")
				if minimalSenderScore > 0 {
					return msmtpd.ErrServiceNotAvailable
				} else {
					return nil
				}
			}
			tr.LogError(err, fmt.Sprintf("while resolving senderscore for transaction address %s", tr.Addr.String()))
			return msmtpd.ErrServiceNotAvailable
		}
		if len(names) == 0 {
			tr.LogInfo("senderscore is 0")
			if minimalSenderScore > 0 {
				return msmtpd.ErrServiceNotAvailable
			} else {
				return nil
			}
		}
		if len(names) > 1 {
			tr.LogError(
				fmt.Errorf("too many responses for senderscore check"),
				fmt.Sprintf("while reversing transaction address %s", tr.Addr.String()))
			return msmtpd.ErrServiceNotAvailable
		}
		if strings.HasPrefix(names[0], "127.0.4.") {
			score := strings.Replace(names[0], "127.0.4.", "", 1)
			senderScore64, parsingErr := strconv.ParseInt(score, 10, 64)
			if parsingErr != nil {
				tr.LogError(
					fmt.Errorf("%s : while parsing %s as senderscore", parsingErr, names[0]),
					fmt.Sprintf("while reversing transaction address %s", tr.Addr.String()))
				return msmtpd.ErrServiceNotAvailable
			}
			tr.LogDebug("SenderScore is %v", senderScore64)
			tr.Incr(SenderscoreCounter, float64(senderScore64))
			if minimalSenderScore > uint(senderScore64) {
				tr.LogInfo("SenderScore %v is lower than %v", senderScore64, minimalSenderScore)
				return msmtpd.ErrServiceNotAvailable
			} else {
				tr.LogInfo("SenderScore %v is bigger than %v", senderScore64, minimalSenderScore)
				return nil
			}
		}
		tr.LogError(
			fmt.Errorf("strange senderscore response %s", names[0]),
			fmt.Sprintf("while reversing transaction address %s", tr.Addr.String()))
		return msmtpd.ErrServiceNotAvailable
	}
}
