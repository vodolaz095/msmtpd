package karma

import (
	"context"

	"msmtpd"
)

// Storage is interface to abstract away saving\retrieving Result with remote IP address karma in it
type Storage interface {
	// Ping ensures Storage works
	Ping(ctx context.Context) error
	// Close closes storage, it should be called before application exits
	Close() error
	// SaveGood saves transaction remote address history as good memory
	SaveGood(*msmtpd.Transaction) error
	// SaveBad saves transaction remote address history as bad memory
	SaveBad(*msmtpd.Transaction) error
	// Get gets karma score for transaction IP address
	Get(*msmtpd.Transaction) (int, error)
}
