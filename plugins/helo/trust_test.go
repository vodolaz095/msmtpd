package helo

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/vodolaz095/msmtpd"
)

func ExampleTrustHellos() {
	trusted := map[string]string{
		"1.1.1.1": "trusted.example.org",
	}

	server := msmtpd.Server{
		Hostname:       "localhost",
		WelcomeMessage: "Do you believe in our God?",
		HeloCheckers: []msmtpd.HelloChecker{
			TrustHellos(trusted),
			SkipHeloCheckForLocal,
			DenyBareIP,
			DenyDynamicIP,
			DenyMalformedDomain,
			DenyReverseDNSMismatch,
		},
	}

	err := server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}

func TestTrustHellos(t *testing.T) {
	cases := []testCase{ //TODO - more and more cases!
		{
			IP:       net.TCPAddr{IP: []byte{1, 1, 1, 1}, Port: 25},
			Helo:     "trusted.example.org",
			ErrorMsg: "",
		},
		{
			IP:       net.TCPAddr{IP: []byte{1, 1, 1, 1}, Port: 25},
			Helo:     "something2.example.org",
			ErrorMsg: msmtpd.ErrServiceDoesNotAcceptEmail.Error(),
		},
		{
			IP:       net.TCPAddr{IP: []byte{1, 1, 1, 2}, Port: 25},
			Helo:     "something2.example.org",
			ErrorMsg: msmtpd.ErrServiceDoesNotAcceptEmail.Error(),
		},
	}
	trusted := map[string]string{
		"1.1.1.1": "trusted.example.org",
	}

	heloTestRunner(t, cases, []msmtpd.HelloChecker{
		TrustHellos(trusted),
		func(_ context.Context, transaction *msmtpd.Transaction) error {
			if transaction.IsFlagSet(IsTrustedOrigin) {
				return nil
			}
			return msmtpd.ErrServiceDoesNotAcceptEmail
		},
	})
}
