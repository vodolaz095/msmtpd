package msmtpd

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"sync"
	"sync/atomic"
	"time"
)

// CheckerFunc are signature of functions used in checks for client issuing HELO/EHLO, MAIL FROM, RCPT TO commands
// First argument is current Transaction, 2nd one - is argument of clients commands.
// Note that we can store counters and Facts in Transaction, in order to extract and reuse it in the future.
type CheckerFunc func(transaction *Transaction) error

// ConnectionChecker are called when tcp connection are established, if they return non-null error, connection is terminated
type ConnectionChecker CheckerFunc

// HelloChecker is called after client provided HELO/EHLO greeting, if they return non-null error, connection is closed
type HelloChecker CheckerFunc

// SenderChecker is called after client provided MAIL FROM, if they return non-null error, connection is closed
type SenderChecker CheckerFunc

// RecipientChecker is called for each RCPT TO client provided, if they return null error, recipient is added to Transaction.RcptTo
type RecipientChecker func(transaction *Transaction, recipient *mail.Address) error

// DataHandler are called when client provided message body, they can be used to either check message by rspamd and header validator, or even actually deliver message to LMTP or 3rd party SMTP server
type DataHandler CheckerFunc

type CloseHandler CheckerFunc

// AuthenticatorFunc is signature of function used to handle authentication
type AuthenticatorFunc func(transaction *Transaction, username, password string) error

// Server defines the parameters for running the SMTP server
type Server struct {
	// Hostname is how we name ourselves, default is "localhost.localdomain"
	Hostname string
	// WelcomeMessage sets initial server banner. (default: "<hostname> ESMTP ready.")
	WelcomeMessage string
	// ReadTimeout is socket timeout for read operations. (default: 60s)
	ReadTimeout time.Duration
	// WriteTimeout is socket timeout for write operations. (default: 60s)
	WriteTimeout time.Duration
	// DataTimeout Socket timeout for DATA command (default: 5m)
	DataTimeout time.Duration

	// MaxConnections sets maximum number of concurrent connections, use -1 to disable. (default: 100)
	MaxConnections int
	// MaxMessageSize, default is
	MaxMessageSize int // Max message size in bytes. (default: 10240000)
	MaxRecipients  int // Max RCPT TO calls for each envelope. (default: 100)

	// Resolver is net.Resolver used by server and plugins to resolve remote resources against DNS servers
	Resolver *net.Resolver

	// Enable various checks during the SMTP session.
	// Can be left empty for no restrictions.
	// If an error is returned, it will be reported in the SMTP session.
	// Use the ErrorSMTP struct for access to error codes.
	// Checks are called synchronously, in usual order

	// ConnectionCheckers are called when TCP connection is started
	ConnectionCheckers []ConnectionChecker
	// HeloCheckers are called after client send HELO/EHLO commands,
	// 1st argument is Transaction, 2nd one - HELO/EHLO payload
	HeloCheckers []HelloChecker
	// SenderCheckers are called when client issues MAIL FROM command,
	// 1st argument is Transaction, 2nd one - MAIL FROM payload
	SenderCheckers []SenderChecker
	// RecipientCheckers are called when client issues RCPT TO command,
	// 1st argument is Transaction, 2nd one - RCPT TO payload
	RecipientCheckers []RecipientChecker

	// Authenticator, while beign not nill, enables PLAIN/LOGIN authentication,
	// only available after STARTTLS. Variable can be left empty for no authentication support.
	Authenticator func(transaction *Transaction, username, password string) error

	// DataHandlers are functions to process message body after DATA command.
	// Can be left empty for a NOOP server.
	// If an error is returned, it will be reported in the SMTP session.
	DataHandlers []DataHandler

	// CloseHandlers are called after connection is closed
	CloseHandlers []CloseHandler

	// EnableXCLIENT enables XClient command support (disabled by default, since it is security risk)
	EnableXCLIENT bool
	// EnableProxyProtocol enables Proxy command support (disabled by default, since it is security risk)
	EnableProxyProtocol bool

	// TLSConfig is used both for STARTTLS and operation over TLS channel
	TLSConfig *tls.Config
	// ForceTLS requires connections to be encrypted
	ForceTLS bool
	// Logger is interface being used as protocol/plugin/errors logger
	Logger Logger

	// mu guards doneChan and makes closing it and listener atomic from
	// perspective of Serve()
	mu         sync.Mutex
	doneChan   chan struct{}
	listener   *net.Listener
	waitgrp    sync.WaitGroup
	inShutdown atomic.Bool
}

// startTransaction takes network connection and wraps it into Transaction object to handle all remote
// client interactions via (E)SMTP protocol.
func (srv *Server) startTransaction(c net.Conn) (t *Transaction) {
	var err error
	id, err := getRandomID()
	if err != nil {
		panic(err) // its extremely unlikely
	}
	mu := sync.Mutex{}
	ctx, cancel := context.WithCancel(context.Background())
	t = &Transaction{
		ID:        id,
		StartedAt: time.Now(),

		server:     srv,
		ServerName: srv.Hostname,
		Logger:     srv.Logger,

		conn:   c,
		reader: bufio.NewReader(c),
		writer: bufio.NewWriter(c),
		Addr:   c.RemoteAddr(),

		ctx:    ctx,
		cancel: cancel,

		Aliases:  nil,
		facts:    make(map[string]string, 0),
		counters: make(map[string]float64, 0),
		flags:    make(map[string]bool, 0),
		mu:       &mu,
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
			t.LogDebug("%s : while performing handshake", err)
			t.Secured = false
			t.Hate(tlsHandshakeFailedHate)
		} else {
			t.Secured = true
		}
		state := tlsConn.ConnectionState()
		t.TLS = &state
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

func (srv *Server) runCloseHandlers(transaction *Transaction) {
	var closeError error
	for k := range srv.CloseHandlers {
		srv.Logger.Debugf(transaction, "Starting close handler %v...", k)
		closeError = srv.CloseHandlers[k](transaction)
		if closeError != nil {
			srv.Logger.Errorf(transaction, "%s : while calling close handler %v", closeError, k)
		} else {
			srv.Logger.Debugf(transaction, "closing handler %v is called", k)
		}
	}
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
		transaction := srv.startTransaction(conn)
		srv.waitgrp.Add(1)
		go func() {
			defer srv.waitgrp.Done()
			defer srv.runCloseHandlers(transaction)
			if limiter != nil {
				select {
				case limiter <- struct{}{}:
					transaction.serve()
					srv.Logger.Infof(transaction, "transaction serving is finished")
					<-limiter
				default:
					transaction.reject()
					srv.Logger.Infof(transaction, "transaction is rejected, server is busy")
				}
			} else {
				transaction.serve()
				srv.Logger.Infof(transaction, "transaction serving is finished")
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
	if srv.Logger == nil {
		srv.Resolver = net.DefaultResolver
	}
	if srv.Logger == nil {
		srv.Logger = &DefaultLogger{
			Logger: log.Default(),
			Level:  InfoLevel,
		}
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
