package procrastinator

import (
	"fmt"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

type testStopWatch struct {
	T         *testing.T
	startedAt time.Time
}

func (tsw *testStopWatch) Start() {
	tsw.startedAt = time.Now()
}

func (tsw *testStopWatch) AtLeast(dur time.Duration) {
	actual := time.Since(tsw.startedAt)
	if actual < dur {
		tsw.T.Errorf("duration %s is lower than expected %s",
			actual.String(), dur.String(),
		)
	} else {
		tsw.T.Logf("Duration is %s", actual.String())
	}

}

func TestProcrastinator(t *testing.T) {
	p := Default()
	tsw := testStopWatch{T: t}
	startedAt := time.Now()
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			p.WaitForConnection(),
		},
		HeloCheckers: []msmtpd.HelloChecker{
			p.WaitForHelo(),
		},
		SenderCheckers: []msmtpd.SenderChecker{
			p.WaitForSender(),
		},
		RecipientCheckers: []msmtpd.RecipientChecker{
			p.WaitForRecipient(),
		},
		DataCheckers: []msmtpd.DataChecker{
			p.WaitForData(),
		},
	})
	defer closer()

	tsw.Start()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	tsw.AtLeast(3 * time.Second)

	tsw.Start()
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	tsw.AtLeast(2 * time.Second)

	tsw.Start()
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	tsw.AtLeast(2 * time.Second)

	tsw.Start()
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	tsw.AtLeast(2 * time.Second)

	tsw.Start()
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	tsw.AtLeast(2 * time.Second)

	tsw.Start()
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprint(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	tsw.AtLeast(2 * time.Second)

	err = c.Quit()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}
	if time.Now().Sub(startedAt) < 6*time.Second {
		t.Errorf("too fast")
	}
}
