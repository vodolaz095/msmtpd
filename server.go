package msmptd

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Server defines the parameters for running the SMTP server
type Server struct {
	Hostname       string // Server hostname. (default: "localhost.localdomain")
	WelcomeMessage string // Initial server banner. (default: "<hostname> ESMTP ready.")

	ReadTimeout  time.Duration // Socket timeout for read operations. (default: 60s)
	WriteTimeout time.Duration // Socket timeout for write operations. (default: 60s)
	DataTimeout  time.Duration // Socket timeout for DATA command (default: 5m)

	MaxConnections int // Max concurrent connections, use -1 to disable. (default: 100)
	MaxMessageSize int // Max message size in bytes. (default: 10240000)
	MaxRecipients  int // Max RCPT TO calls for each envelope. (default: 100)

	// New e-mails are handed off to this functions.
	// Can be left empty for a NOOP server.
	// If an error is returned, it will be reported in the SMTP session.
	Handlers []func(transaction *Transaction) error

	// Enable various checks during the SMTP session.
	// Can be left empty for no restrictions.
	// If an error is returned, it will be reported in the SMTP session.
	// Use the ErrorSMTP struct for access to error codes.
	ConnectionCheckers []func(transaction *Transaction) error              // Called upon new connection.
	HeloCheckers       []func(transaction *Transaction, name string) error // Called after HELO/EHLO.
	SenderCheckers     []func(transaction *Transaction, addr string) error // Called after MAIL FROM.
	RecipientCheckers  []func(transaction *Transaction, addr string) error // Called after each RCPT TO.

	// Enable PLAIN/LOGIN authentication, only available after STARTTLS.
	// Can be left empty for no authentication support.
	Authenticator func(transaction *Transaction, username, password string) error

	EnableXCLIENT       bool // Enable XCLIENT support (default: false)
	EnableProxyProtocol bool // Enable proxy protocol support (default: false)

	TLSConfig *tls.Config // Enable STARTTLS support.
	ForceTLS  bool        // Force STARTTLS usage.

	Logger Logger

	// mu guards doneChan and makes closing it and listener atomic from
	// perspective of Serve()
	mu         sync.Mutex
	doneChan   chan struct{}
	listener   *net.Listener
	waitgrp    sync.WaitGroup
	inShutdown atomic.Bool
}

func (srv *Server) newSession(c net.Conn) (t *Transaction) {
	var err error
	t = &Transaction{
		server:     srv,
		conn:       c,
		reader:     bufio.NewReader(c),
		writer:     bufio.NewWriter(c),
		Addr:       c.RemoteAddr(),
		ServerName: srv.Hostname,
		facts:      make(map[string]string, 0),
		counters:   make(map[string]float64, 0),
	}

	// Check if the underlying connection is already TLS.
	// This will happen if the Listener provided Serve()
	// is from tls.Listen()
	var tlsConn *tls.Conn
	tlsConn, t.Encrypted = c.(*tls.Conn)
	if t.Encrypted {
		// run handshake otherwise it's done when we first
		// read/write and connection state will be invalid
		err = tlsConn.Handshake()
		if err != nil {
			t.Secured = false
		} else {
			t.Secured = true
		}
		state := tlsConn.ConnectionState()
		t.TLS = &state
		t.Hate(tlsHandshakeFailedHate)
	}
	t.scanner = bufio.NewScanner(t.reader)
	return
}

// ListenAndServe starts the SMTP server and listens on the address provided
func (srv *Server) ListenAndServe(addr string) error {
	if srv.inShutdown.Load() {
		return ErrServerClosed
	}

	srv.configureDefaults()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return srv.Serve(l)
}

// Serve starts the SMTP server and listens on the Listener provided
func (srv *Server) Serve(l net.Listener) error {
	if srv.inShutdown.Load() {
		return ErrServerClosed
	}

	srv.configureDefaults()

	l = &onceCloseListener{Listener: l}
	defer l.Close()
	srv.listener = &l

	var limiter chan struct{}

	if srv.MaxConnections > 0 {
		limiter = make(chan struct{}, srv.MaxConnections)
	}

	for {
		conn, e := l.Accept()
		if e != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}

			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				time.Sleep(time.Second)
				continue
			}
			return e
		}

		session := srv.newSession(conn)

		srv.waitgrp.Add(1)
		go func() {
			defer srv.waitgrp.Done()
			if limiter != nil {
				select {
				case limiter <- struct{}{}:
					session.serve()
					<-limiter
				default:
					session.reject()
				}
			} else {
				session.serve()
			}
		}()
	}

}

// Shutdown instructs the server to shut down, starting by closing the
// associated listener. If wait is true, it will wait for the shutdown
// to complete. If wait is false, Wait must be called afterwards.
func (srv *Server) Shutdown(wait bool) error {
	var lnerr error
	srv.inShutdown.Store(true)

	// First close the listener
	srv.mu.Lock()
	if srv.listener != nil {
		lnerr = (*srv.listener).Close()
	}
	srv.closeDoneChanLocked()
	srv.mu.Unlock()

	// Now wait for all client connections to close
	if wait {
		srv.Wait()
	}

	return lnerr
}

// Wait waits for all client connections to close and the server to finish
// shutting down.
func (srv *Server) Wait() error {
	if !srv.inShutdown.Load() {
		return errors.New("server has not been shutdown")
	}
	srv.waitgrp.Wait()
	return nil
}

// Address returns the listening address of the server
func (srv *Server) Address() net.Addr {
	return (*srv.listener).Addr()
}

func (srv *Server) configureDefaults() {
	if srv.MaxMessageSize == 0 {
		srv.MaxMessageSize = 10240000
	}
	if srv.MaxConnections == 0 {
		srv.MaxConnections = 100
	}
	if srv.MaxRecipients == 0 {
		srv.MaxRecipients = 100
	}
	if srv.ReadTimeout == 0 {
		srv.ReadTimeout = time.Second * 60
	}
	if srv.WriteTimeout == 0 {
		srv.WriteTimeout = time.Second * 60
	}
	if srv.DataTimeout == 0 {
		srv.DataTimeout = time.Minute * 5
	}
	if srv.ForceTLS && srv.TLSConfig == nil {
		log.Fatal("Cannot use ForceTLS with no TLSConfig")
	}
	if srv.Hostname == "" {
		srv.Hostname = "localhost.localdomain"
	}
	if srv.WelcomeMessage == "" {
		srv.WelcomeMessage = fmt.Sprintf("%s ESMTP ready.", srv.Hostname)
	}
}

// From net/http/server.go

func (srv *Server) shuttingDown() bool {
	return srv.inShutdown.Load()
}

func (srv *Server) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.getDoneChanLocked()
}

func (srv *Server) getDoneChanLocked() chan struct{} {
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *Server) closeDoneChanLocked() {
	ch := srv.getDoneChanLocked()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		// by s.mu.
		close(ch)
	}
}

// onceCloseListener wraps a net.Listener, protecting it from
// multiple Close calls.
type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

// Close closes
func (oc *onceCloseListener) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (oc *onceCloseListener) close() { oc.closeErr = oc.Listener.Close() }
