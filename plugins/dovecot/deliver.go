package dovecot

import (
	"context"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/plugins/deliver"
)

// Deliver actually delivers message into dovecot server LMTP socket
func (d *Dovecot) Deliver(ctx context.Context, tr *msmtpd.Transaction) (err error) {
	opts := deliver.LMTPOptions{
		Network: "unix",
		Address: d.LmtpSocket,
		LHLO:    "localhost",
		Timeout: d.Timeout,
	}
	return deliver.ViaLocalMailTransferProtocol(opts)(ctx, tr)
}
