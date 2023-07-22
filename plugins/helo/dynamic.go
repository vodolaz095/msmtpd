package helo

import (
	"msmtpd"
)

func isDynamic(tr *msmtpd.Transaction) bool {
	// raw := tr.Addr.(*net.TCPAddr).IP.MarshalText()

	// examples of dynamic ptrs (dig -x {ipAddr})

	//tr.Addr.String()
	//"192.0.2.1:25", "[2001:db8::1]:80")

	// 193.41.76.171 =>
	// r193-41-76-171.utex-telecom.ru.

	// 2606:4700:1101:2:b76d:d11a:e187:e33
	// 0.0.7.4.6.0.6.2.ip6.arpa.

	return false
}
