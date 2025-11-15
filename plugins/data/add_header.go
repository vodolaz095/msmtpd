package data

import (
	"context"

	"github.com/vodolaz095/msmtpd"
)

// AddHeader adds header to both body and parsed headers of msmtpd.Transaction
func AddHeader(name, value string) msmtpd.DataChecker {
	return func(_ context.Context, tr *msmtpd.Transaction) error {
		tr.AddHeader(name, value)
		return nil
	}
}
