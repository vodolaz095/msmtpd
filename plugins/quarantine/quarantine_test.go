package quarantine

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

func TestQuarantine(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "msmptd")
	var tID string
	var createdAt time.Time
	validMessage := internal.MakeTestMessage("scuba@vodolaz095.ru", "scuba@vodolaz095.ru")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.SetFlag(FlagName)
				tID = tr.ID
				createdAt = tr.StartedAt
				return nil
			},
		},
		DataHandlers: []msmtpd.DataHandler{
			MoveToDirectory(dir),
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = c.Mail("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("scuba@vodolaz095.ru"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprint(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(
		dir,
		createdAt.Format("2006"),
		createdAt.Format("01"),
		createdAt.Format("02"),
		tID+".eml",
	))
	if err != nil {
		t.Errorf("%s : while reading file", err)
	}
	t.Logf("Quarantined message is %s", string(data))
	err = os.RemoveAll(dir)
	if err != nil {
		t.Errorf("%s : while removing temporary directory %s", err, dir)
	}
}
