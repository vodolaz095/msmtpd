package dovecot

import (
	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/plugins/deliver"
)

// https://en.wikipedia.org/wiki/Local_Mail_Transfer_Protocol
// https://github.com/emersion/go-smtp/blob/master/lmtp_server_test.go
// https://www.rfc-editor.org/rfc/rfc2033.html#section-4.2

// Deliver actually delivers message into dovecot server LMTP socket
func (d *Dovecot) Deliver(tr *msmtpd.Transaction) (err error) {
	opts := deliver.LMTPOptions{
		Network: "unix",
		Address: d.LmtpSocket,
		LHLO:    "localhost",
		Timeout: d.Timeout,
	}
	return deliver.ViaLocalMailTransferProtocol(opts)(tr)
}
