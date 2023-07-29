package procrastinator

import (
	"crypto/rand"
	"math/big"
	"net/mail"
	"time"

	"github.com/vodolaz095/msmtpd"
)

// Procrastinator used to add random delays based on transaction karma - lower the karma, more the delays
type Procrastinator struct {
	// ConstantDelay added to all calls
	ConstantDelay time.Duration
	// RandomDelay can be used to decrease efficiency of timing attacks for bruteforcing passwords
	RandomDelay time.Duration
	// KarmaCoefficient makes things go faster when transaction karma is good
	KarmaCoefficient time.Duration
}

// Default makes Procrastinator with sane default values
func Default() Procrastinator {
	return Procrastinator{
		ConstantDelay:    3 * time.Second,
		RandomDelay:      time.Second,
		KarmaCoefficient: 100 * time.Millisecond,
	}
}

func doRandomDelay(dur time.Duration) (howMuch time.Duration, err error) {
	max := big.NewInt(dur.Milliseconds())
	delay, err := rand.Int(rand.Reader, max)
	if err != nil {
		return
	}
	return time.Duration(delay.Int64()) * time.Millisecond, nil
}

// wait should be called when you want client to train patience
func (p *Procrastinator) wait() msmtpd.CheckerFunc {
	return func(t *msmtpd.Transaction) (err error) {
		var randomDelay time.Duration
		if p.RandomDelay != 0 {
			randomDelay, err = doRandomDelay(p.RandomDelay)
			if err != nil {
				t.LogError(err, "while getting random delay")
				return msmtpd.ErrServiceNotAvailable
			}
		}
		delay := p.ConstantDelay - time.Duration(t.Karma())*p.KarmaCoefficient + randomDelay
		t.LogDebug("Waiting %s", delay.String())
		time.Sleep(delay)
		return nil
	}
}

// WaitForConnection should be called when you want client to train patience waiting when server will greet you
func (p *Procrastinator) WaitForConnection() msmtpd.ConnectionChecker {
	return func(tr *msmtpd.Transaction) error {
		return p.wait()(tr)
	}
}

// WaitForHelo should be called when you want client to train patience waiting for HELO/EHLO
func (p *Procrastinator) WaitForHelo() msmtpd.HelloChecker {
	return func(tr *msmtpd.Transaction) error {
		return p.wait()(tr)
	}
}

// WaitForSender should be called when you want client to train patience waiting for MAIL FROM
func (p *Procrastinator) WaitForSender() msmtpd.SenderChecker {
	return func(tr *msmtpd.Transaction) error {
		return p.wait()(tr)
	}
}

// WaitForRecipient should be called when you want client to train patience waiting for RCPT TO
func (p *Procrastinator) WaitForRecipient() msmtpd.RecipientChecker {
	return func(tr *msmtpd.Transaction, recipient *mail.Address) error {
		return p.wait()(tr)
	}
}

// WaitForData should be called when you want client to train waiting for DATA
func (p *Procrastinator) WaitForData() msmtpd.DataChecker {
	return func(tr *msmtpd.Transaction) error {
		return p.wait()(tr)
	}
}
