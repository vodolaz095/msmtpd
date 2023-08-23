package redis

import (
	"context"
	"fmt"
	"net"

	"github.com/redis/go-redis/v9"
	"github.com/vodolaz095/msmtpd"
)

// Storage saves IP address history into redis database
type Storage struct {
	Client *redis.Client
}

// Ping tests connection to redis database
func (s *Storage) Ping(ctx context.Context) error {
	return s.Client.Ping(ctx).Err()
}

// Close closes
func (s *Storage) Close() error {
	return s.Client.Close()
}

func (s *Storage) getKey(transaction *msmtpd.Transaction) string {
	return fmt.Sprintf("karma|%s", transaction.Addr.(*net.TCPAddr).IP.String())
}

// SaveGood saves transaction signature as good
func (s *Storage) SaveGood(transaction *msmtpd.Transaction) (err error) {
	key := s.getKey(transaction)
	err = s.Client.HIncrBy(transaction.Context(), key, "connections", 1).Err()
	if err != nil {
		return
	}
	return s.Client.HIncrBy(transaction.Context(), key, "good", 1).Err()
}

// SaveBad saves transaction signature as bad
func (s *Storage) SaveBad(transaction *msmtpd.Transaction) (err error) {
	key := s.getKey(transaction)
	err = s.Client.HIncrBy(transaction.Context(), key, "connections", 1).Err()
	if err != nil {
		return
	}
	return s.Client.HIncrBy(transaction.Context(), key, "bad", 1).Err()
}

// Score used to pack IP address history in memory
type Score struct {
	Connections uint `redis:"connections"`
	Good        uint `redis:"good"`
	Bad         uint `redis:"bad"`
}

// Get extracts transaction karma score
func (s *Storage) Get(transaction *msmtpd.Transaction) (int, error) {
	key := s.getKey(transaction)
	var score Score
	err := s.Client.HMGet(transaction.Context(), key, "connections", "good", "bad").Scan(&score)
	if err != nil {
		if err != redis.Nil {
			return 0, err
		}
		return 0, nil
	}
	return int(score.Good) - int(score.Bad), nil
}

// haraka format is
//
// key - karma|65.49.20.88
// keytype hash
// connections - 4
// good - 3
// bad - 1
