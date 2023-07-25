package internal

import (
	"crypto/tls"
	"net"
	"testing"
)

type TestServer interface {
	Address() net.Addr
	Serve(listener net.Listener) error
}

func RunServerWithoutTLS(t *testing.T, server TestServer) (addr string, closer func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("Listen failed: %v", err)
	}
	go func() {
		serveErr := server.Serve(ln)
		if err != nil {
			t.Errorf("%s : while starting server on %s",
				serveErr, server.Address())
		}
	}()
	done := make(chan bool)
	go func() {
		<-done
		ln.Close()
	}()
	return ln.Addr().String(), func() {
		done <- true
	}
}

func MakeTLSForLocalhost() (cfg *tls.Config, err error) {
	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	if err != nil {
		return
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
