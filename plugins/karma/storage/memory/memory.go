package memory

import (
	"context"
	"net"
	"sync"

	"github.com/vodolaz095/msmtpd"
)

// Score used to pack IP address history in memory
type Score struct {
	Good        uint
	Bad         uint
	Connections uint
}

// Storage saves IP address history in memory
type Storage struct {
	mu   sync.RWMutex
	Data map[string]Score
}

// Ping does nothing, but somehow prepares memory storage
func (m *Storage) Ping(ctx context.Context) error {
	if m.Data == nil {
		m.Data = make(map[string]Score, 0)
	}
	return nil
}

// Close purges memory storage
func (m *Storage) Close() error {
	m.Data = nil
	return nil
}

// SaveGood saves transaction remote address history as good memory
func (m *Storage) SaveGood(transaction *msmtpd.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := transaction.Addr.(*net.TCPAddr).IP.String()
	old, found := m.Data[key]
	if found {
		old.Good++
		old.Connections++
	} else {
		old = Score{Good: 1, Bad: 0, Connections: 1}
	}
	m.Data[key] = old
	return nil
}

// SaveBad saves transaction remote address history as bad memory
func (m *Storage) SaveBad(transaction *msmtpd.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := transaction.Addr.(*net.TCPAddr).IP.String()
	old, found := m.Data[key]
	if found {
		old.Bad++
		old.Connections++
	} else {
		old = Score{Good: 0, Bad: 1, Connections: 1}
	}
	m.Data[key] = old
	return nil
}

// Get gets karma score for transaction IP address
func (m *Storage) Get(transaction *msmtpd.Transaction) (int, error) {
	key := transaction.Addr.(*net.TCPAddr).IP.String()
	m.mu.RLock()
	defer m.mu.RUnlock()
	old, found := m.Data[key]
	if found {
		return int(old.Good) - int(old.Bad), nil
	}
	return 0, nil
}
