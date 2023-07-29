package file

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"

	"github.com/vodolaz095/msmtpd"
)

// Data used to pack IP address history in file
type Data struct {
	Connections uint `json:"connections"`
	Good        uint `json:"good"`
	Bad         uint `json:"bad"`
}

// Storage saves IP address history into files
type Storage struct {
	Directory string
}

// Ping pretends it does anything useful
func (f *Storage) Ping(ctx context.Context) error {
	return nil
}

// Close closes
func (f *Storage) Close() error {
	return nil
}

func (f *Storage) getFileName(transaction *msmtpd.Transaction) string {
	key := transaction.Addr.(*net.TCPAddr).IP.String()
	return filepath.Join(f.Directory, key+".json")
}

func (f *Storage) loadTransactionData(transaction *msmtpd.Transaction) (data Data, err error) {
	contents, err := os.ReadFile(f.getFileName(transaction))
	if err != nil {
		if os.IsNotExist(err) {
			return Data{
				Connections: 0,
				Good:        0,
				Bad:         0,
			}, nil
		}
		return
	}
	err = json.Unmarshal(contents, &data)
	return
}

func (f *Storage) saveTransactionData(transaction *msmtpd.Transaction, data Data) (err error) {
	bdy, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return
	}
	h, err := os.OpenFile(f.getFileName(transaction), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	_, err = h.Write(bdy)
	if err != nil {
		return
	}
	return h.Close()
}

// SaveGood saves transaction remote address history as good memory
func (f *Storage) SaveGood(transaction *msmtpd.Transaction) (err error) {
	data, err := f.loadTransactionData(transaction)
	if err != nil {
		return
	}
	data.Good++
	data.Connections++
	return f.saveTransactionData(transaction, data)
}

// SaveBad saves transaction remote address history as bad memory
func (f *Storage) SaveBad(transaction *msmtpd.Transaction) (err error) {
	data, err := f.loadTransactionData(transaction)
	if err != nil {
		return
	}
	data.Bad++
	data.Connections++
	return f.saveTransactionData(transaction, data)
}

// Get gets karma score for transaction IP address
func (f *Storage) Get(transaction *msmtpd.Transaction) (int, error) {
	data, err := f.loadTransactionData(transaction)
	if err != nil {
		return 0, err
	}
	return int(data.Good) - int(data.Bad), nil
}
