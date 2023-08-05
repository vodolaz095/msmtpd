package msmtpd

import (
	"crypto/tls"
	"net"
	"net/smtp"
	"testing"

	"github.com/vodolaz095/msmtpd/internal"
)

func TestStartWithTLS(t *testing.T) {
	cfg, err := internal.MakeTLSForLocalhost()
	if err != nil {
		t.Fatalf("%s : while loading test certs for localhost", err)
	}
	clear, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("%s : while starting clear text listener", err)
	}
	server := Server{
		HeloCheckers: []HelloChecker{
			func(tr *Transaction) error {
				if !tr.Encrypted {
					t.Errorf("connection is not encrypted")
				}
				if !tr.Secured {
					t.Errorf("connection is not secured properly")
				}
				return nil
			},
		},
	}
	server.TLSConfig = cfg
	logger := TestLogger{Suite: t}
	server.Logger = &logger
	ln := tls.NewListener(clear, cfg)
	server.configureDefaults()

	go func() {
		server.Serve(ln)
	}()
	done := make(chan bool)
	go func() {
		<-done
		ln.Close()
	}()
	conn, err := tls.Dial("tcp", ln.Addr().String(), &tls.Config{
		ServerName:         "localhost", // SNI :-)
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,
	})

	c, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		t.Errorf("%s : while connecting to %s", err, ln.Addr().String())
	}
	err = c.Hello("localhost")
	if err != nil {
		t.Errorf("%s : while making helo", err)
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while making helo", err)
	}
	done <- true
}
