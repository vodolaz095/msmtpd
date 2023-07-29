package file

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func TestStorage(t *testing.T) {
	var score int
	dir := filepath.Join(os.TempDir(), "test_karma_storage")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("%s : while creating temp karma directory at %s", err, dir)
	}
	storage := Storage{Directory: dir}
	err = storage.Ping(context.TODO())
	if err != nil {
		t.Errorf("%s : while pinging storage", err)
	}
	tr := msmtpd.Transaction{
		Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.3"), Port: 25},
	}
	err = storage.SaveGood(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as good", err)
	}
	score, err = storage.Get(&tr)
	if err != nil {
		t.Errorf("%s : while geting transaction", err)
	}
	t.Logf("Score %v", score)
	if score != 1 {
		t.Errorf("wrong score %v instead of 1", score)
	}
	err = storage.SaveBad(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as bad", err)
	}
	score, err = storage.Get(&tr)
	if err != nil {
		t.Errorf("%s : while geting transaction", err)
	}
	t.Logf("Score %v", score)
	if score != 0 {
		t.Errorf("wrong score %v instead of 0", score)
	}
	err = storage.Close()
	if err != nil {
		t.Errorf("%s : while closing storage", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "192.168.1.3.json"))
	if err != nil {
		t.Errorf("%s : while reading file", err)
	}
	t.Logf("Data: %s", string(data))
	err = os.RemoveAll(dir)
	if err != nil {
		t.Errorf("%s : while purging data", err)
	}

}
