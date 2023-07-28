package memory

import (
	"context"
	"net"
	"testing"

	"msmtpd"
)

func TestStorage(t *testing.T) {
	var score int
	var err error
	storage := Storage{}
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
	err = storage.SaveGood(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as good", err)
	}
	score, err = storage.Get(&tr)
	if err != nil {
		t.Errorf("%s : while geting transaction", err)
	}
	t.Logf("Score %v", score)
	if score != 2 {
		t.Errorf("wrong score %v instead of 2", score)
	}
	err = storage.SaveBad(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as bad", err)
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
	raw, found := storage.Data["192.168.1.3"]
	if !found {
		t.Errorf("key is not found")
	} else {
		t.Logf("raw: %v", raw)
		if raw.Connections != 4 {
			t.Errorf("Wrong connections %v, 4 expected", raw.Connections)
		}
		if raw.Bad != 2 {
			t.Errorf("Wrong bad connections %v, 2 expected", raw.Bad)
		}
		if raw.Good != 2 {
			t.Errorf("Wrong good connections %v, 2 expected", raw.Good)
		}
	}
	err = storage.Close()
	if err != nil {
		t.Errorf("%s : while closing storage", err)
	}
}
