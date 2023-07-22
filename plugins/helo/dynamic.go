package helo

import "msmtpd"

func isDynamic(tr *msmtpd.Transaction) bool {
	// examples of dynamic ptrs (dig -x {ipAddr})

	// 193.41.76.171 =>
	// r193-41-76-171.utex-telecom.ru.

	// 2606:4700:1101:2:b76d:d11a:e187:e33
	// 0.0.7.4.6.0.6.2.ip6.arpa.

	return false
}
