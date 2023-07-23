package karma

import (
	"fmt"
	"net/smtp"
	"sync"
	"testing"

	"msmtpd"
	"msmtpd/plugins/karma/storage/memory"
)

func TestKarmaPluginMemoryGood(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	memStorage := memory.Storage{Data: make(map[string]memory.Score, 0)}
	memStorage.Data["127.0.0.1"] = memory.Score{
		Good:        5,
		Bad:         0,
		Connections: 5,
	}
	kh := Handler{
		HateLimit: 4,
		Storage:   &memStorage,
	}

	addr, closer := runserver(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			kh.ConnectionChecker,
		},
		CloseHandlers: []msmtpd.CloseHandler{
			kh.CloseHandler,
			func(transaction *msmtpd.Transaction) error {
				wg.Done()
				return nil
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("RCPT failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Error("while closing email body stream")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("QUIT failed: %v", err)
	}
	wg.Wait()
	score, found := memStorage.Data["127.0.0.1"]
	if !found {
		t.Errorf("karma is nuked?")
	}
	t.Logf("Score: %v connections, %v good and %v bad", score.Connections, score.Good, score.Bad)
	if score.Connections != 6 {
		t.Errorf("wrong connections %v isntead 6", score.Connections)
	}
	if score.Good != 6 {
		t.Errorf("wrong good connecetions %v isntead of 6", score.Good)
	}
	if score.Bad != 0 {
		t.Errorf("wrong bad connecetions %v isntead of 0", score.Bad)
	}
}

func TestKarmaPluginMemoryBad(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	memStorage := memory.Storage{Data: make(map[string]memory.Score, 0)}
	memStorage.Data["127.0.0.1"] = memory.Score{
		Good:        0,
		Bad:         5,
		Connections: 5,
	}
	kh := Handler{
		HateLimit: 5,
		Storage:   &memStorage,
	}

	addr, closer := runserver(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			kh.ConnectionChecker,
		},
		CloseHandlers: []msmtpd.CloseHandler{
			kh.CloseHandler,
			func(transaction *msmtpd.Transaction) error {
				wg.Done()
				return nil
			},
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("wrong error %s", err)
		}
	} else {
		t.Errorf("we are not banned???")
	}
	wg.Wait()
	score, found := memStorage.Data["127.0.0.1"]
	if !found {
		t.Errorf("karma is nuked?")
	}
	t.Logf("Score: %v connections, %v good and %v bad", score.Connections, score.Good, score.Bad)
	if score.Connections != 6 {
		t.Errorf("wrong connections %v isntead 6", score.Connections)
	}
	if score.Good != 0 {
		t.Errorf("wrong good connecetions %v isntead of 0", score.Good)
	}
	if score.Bad != 6 {
		t.Errorf("wrong bad connecetions %v isntead of 6", score.Bad)
	}
}
