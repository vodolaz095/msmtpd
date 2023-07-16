package dovecot

import (
	"net/mail"
	"os"
	"testing"
	"time"

	"msmtpd"
)

var username, password, rcptTo string

func TestLoadFromEnvironment(t *testing.T) {
	username = os.Getenv("DOVECOT_USERNAME")
	if username == "" {
		t.Errorf("environment variable DOVECOT_USERNAME is not set")
	}
	password = os.Getenv("DOVECOT_PASSWORD")
	if password == "" {
		t.Errorf("environment variable DOVECOT_PASSWORD is not set")
	}
	rcptTo = os.Getenv("DOVECOT_RCPT_TO")
	if rcptTo == "" {
		t.Errorf("environment variable DOVECOT_RCPT_TO is not set")
	}
}

func TestDovecot_Exists(t *testing.T) {
	if rcptTo == "" {
		t.Skipf("skipping, because environment variable DOVECOT_RCPT_TO is not set")
	}

	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LtmpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	tr := msmtpd.Transaction{
		ID:        "dovecot_exists",
		StartedAt: time.Now(),
	}
	err := dvc.Exists(&tr, &mail.Address{
		Name:    "who cares",
		Address: rcptTo,
	})
	if err != nil {
		t.Errorf("%s : while checking %s to exists", err, rcptTo)
	}
	err = dvc.Exists(&tr, &mail.Address{
		Name:    "who cares",
		Address: "somebody@example.org",
	})
	if err != nil {
		if err.Error() != "521 i have no idea about recipient you want to deliver message to" {
			t.Errorf("%s : wrong error while checking non existant mailbox", err)
		}
	} else {
		t.Errorf("error is expected")
	}
}

func TestDovecot_Authenticate(t *testing.T) {
	if username == "" {
		t.Skipf("skipping, because environment variable DOVECOT_USERNAME is not set")
	}
	if password == "" {
		t.Skipf("skipping, because environment variable DOVECOT_PASSWORD is not set")
	}

	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LtmpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	tr := msmtpd.Transaction{
		ID:        "dovecot_authenticate",
		StartedAt: time.Now(),
	}
	err := dvc.Authenticate(&tr, username, password)
	if err != nil {
		t.Errorf("%s : while checking %s to exists", err, rcptTo)
	}
	err = dvc.Authenticate(&tr, "wrong_username", "wrong_password")
	if err != nil {
		if err.Error() != "521 authorization failed" {
			t.Errorf("%s : wrong error while checking wrong authentication", err)
		}
	} else {
		t.Errorf("error is expected")
	}
}

func TestDovecot_Deliver(t *testing.T) {
	validMessage := `Date: Sun, 11 Jun 2023 19:49:29 +0300
To: scuba@vodolaz095.ru
From: scuba@vodolaz095.ru
Subject: test Sun, 11 Jun 2023 19:49:29 +0300
Message-Id: <20230611194929.017435@localhost>
X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/

This is a test mailing during dovecot unit test
`

	if rcptTo == "" {
		t.Skipf("skipping, because environment variable DOVECOT_RCPT_TO is not set")
	}
	dvc := Dovecot{
		PathToAuthUserDBSocket: DefaultAuthUserSocketPath,
		PathToAuthClientSocket: DefaultClientSocketPath,
		LtmpSocket:             DefaultLMTPSocketPath,
		Timeout:                5 * time.Second,
	}
	tr := msmtpd.Transaction{
		ID:        "dovecot_deliver",
		StartedAt: time.Now(),
		Body:      []byte(validMessage),
		MailFrom:  mail.Address{Name: "who cares", Address: rcptTo},
		RcptTo: []mail.Address{
			{Name: "who cares", Address: rcptTo},
		},
	}
	err := dvc.Deliver(&tr)
	if err != nil {
		t.Errorf("%s : while delivering test message", err)
	}
}
