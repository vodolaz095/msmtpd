package karma

import (
	"context"

	"msmtpd"
)

type Result struct {
	Connections uint
	Good        uint
	Bad         uint
}

type Storage interface {
	Ping(ctx context.Context) error
	Close() error
	SaveGood(*msmtpd.Transaction) error
	SaveBad(*msmtpd.Transaction) error
	Get(*msmtpd.Transaction) (int, error)
}

// TODO - for redis
//
// key - karma|65.49.20.88
// keytype hash
// connections - 4
// good - 3
// bad - 1
