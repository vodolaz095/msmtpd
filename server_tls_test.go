package msmtpd

import (
	"crypto/tls"
	"net"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd/internal"
)

func TestStartWithTLS_OK(t *testing.T) {
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

func TestStartWithTLS_SNI_fail(t *testing.T) {
	cfg, err := internal.MakeTLSForLocalhost()
	if err != nil {
		t.Fatalf("%s : while loading test certs for localhost", err)
	}
	clear, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("%s : while starting clear text listener", err)
	}
	server := Server{}
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
	_, err = tls.Dial("tcp", ln.Addr().String(), &tls.Config{
		ServerName:         "smtp.yandex.ru", // SNI will fail
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: false,
	})
	if err != nil {
		t.Logf("tls handshake error: %s", err)
		if err.Error() != "x509: certificate is not valid for any names, but wanted to match smtp.yandex.ru" {
			t.Errorf("%s : while making TLS connection", err)
		}
	} else {
		t.Errorf("tls connection error not thrown")
	}
	time.Sleep(time.Second)
	done <- true
}

func TestStartWithTLS_SNI_unsecure(t *testing.T) {
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
				t.Log("Handshake complete", tr.TLS.HandshakeComplete)
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
		ServerName:         "smtp.yandex.ru", // SNI will fail
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // never do it in production
	})
	if err != nil {
		t.Errorf("%s : while making TLS connection", err)
	}
	c, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		t.Errorf("%s : while making smtp client over TLS connection", err)
	}
	err = c.Hello("localhost")
	if err != nil {
		t.Errorf("%s : while making HELO over TLS connection", err)
	}

	time.Sleep(time.Second)
	done <- true
}
