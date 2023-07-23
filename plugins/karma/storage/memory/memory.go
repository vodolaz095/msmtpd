package memory

import (
	"context"
	"net"
	"sync"

	"msmtpd"
)

type Score struct {
	Good        uint
	Bad         uint
	Connections uint
}

type Storage struct {
	mu   sync.RWMutex
	Data map[string]Score
}

func (m *Storage) Ping(ctx context.Context) error {
	return nil
}

func (m *Storage) Close() error {
	m.Data = nil
	return nil
}

func (m *Storage) SaveGood(transaction *msmtpd.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := transaction.Addr.(*net.TCPAddr).IP.String()
	old, found := m.Data[key]
	if found {
		old.Good += 1
		old.Connections += 1
	} else {
		old = Score{Good: 1, Bad: 0, Connections: 1}
	}
	m.Data[key] = old
	return nil
}

func (m *Storage) SaveBad(transaction *msmtpd.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := transaction.Addr.(*net.TCPAddr).IP.String()
	old, found := m.Data[key]
	if found {
		old.Bad += 1
		old.Connections += 1
	} else {
		old = Score{Good: 0, Bad: 1, Connections: 1}
	}
	m.Data[key] = old
	return nil
}

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
