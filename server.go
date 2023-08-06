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

	"go.opentelemetry.io/otel/attribute"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// CheckerFunc are signature of functions used in checks for client issuing HELO/EHLO, MAIL FROM, RCPT TO commands
// Note that we can store counters and Facts in Transaction, in order to extract and reuse it in the future.
type CheckerFunc func(transaction *Transaction) error

// ConnectionChecker are called when tcp connection are established, if they return non-null error,
// connection is terminated
type ConnectionChecker CheckerFunc

// HelloChecker is called after client provided HELO/EHLO greeting, returned errors are send
// to client as ErrorSMTP responses
type HelloChecker CheckerFunc

// SenderChecker is called after client provided MAIL FROM, returned errors are send
// // to client as ErrorSMTP responses
type SenderChecker CheckerFunc

// RecipientChecker is called for each RCPT TO client provided, if they return null error,
// recipient is added to Transaction.RcptTo, else returned errors are send
// to client as ErrorSMTP responses
type RecipientChecker func(transaction *Transaction, recipient *mail.Address) error

// DataChecker is called when client provided message body, and we need to ensure it is sane.
// It is good place to use RSPAMD and other message body validators here
type DataChecker CheckerFunc

// DataHandler is called when client provided message body, it was checked by all DataChecker functions,
// and we need to deliver message to LMTP or 3rd party SMTP server.
type DataHandler CheckerFunc

// CloseHandler is called when server terminates SMTP session, it can be used for,
// for example, storing Karma or reporting statistics
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
	// MaxMessageSize, default is 10240000 bytes
	MaxMessageSize int
	// MaxRecipients are limit for RCPT TO calls for each envelope. (default: 100)
	MaxRecipients int

	// Resolver is net.Resolver used by server and plugins to resolve remote resources against DNS servers
	Resolver *net.Resolver

	// SkipResolvingPTR disables resolving reverse/point DNS records of connecting IP address,
	// it can be usefull in various DNS checks, but it reduces perfomance due to quite
	// expensive and slow DNS calls. By default resolving PTR records is enabled
	SkipResolvingPTR bool

	// Enable various checks during the SMTP session.
	// Can be left empty for no restrictions.
	// If an error is returned, it will be reported in the SMTP session.
	// Use the ErrorSMTP struct for access to error codes.
	// Checks are called synchronously, in usual order

	// ConnectionCheckers are called when TCP connection is started, if any of connection
	// checkers returns error, connection is closed, which is reported as ErrorSMTP being send to
	// client before connection is terminated
	ConnectionCheckers []ConnectionChecker

	// HeloCheckers are called after client send HELO/EHLO commands
	// If any of HeloCheckers returns error, it will be reported as HELO/EHLO command response,
	// and HELO command will be considered erroneous.
	HeloCheckers []HelloChecker

	// SenderCheckers are called when client issues MAIL FROM command
	// If any of SenderCheckers returns error, it will be reported as MAIL FROM command response,
	// and command will be considered erroneous.
	SenderCheckers []SenderChecker

	// RecipientCheckers are called every time client issues RCPT TO command
	// 1st argument is Transaction, 2nd one - RCPT TO payload
	// If any of RecipientCheckers returns error, it will be reported as RCPT TO command response
	// and command will be considered erroneous.
	RecipientCheckers []RecipientChecker

	// Authenticator, while being not nil, enables PLAIN/LOGIN authentication,
	// only available after STARTTLS. Variable can be left empty for no authentication support.
	// If Authenticator returns error, authentication will be considered erroneous.
	Authenticator func(transaction *Transaction, username, password string) error

	// DataCheckers are functions called to check message body before passing it
	// to DataHandlers for delivery. If left empty, body is not checked. It is worth
	// mentioning that message body is parsed according to RFC 5322 to ensure mandatory
	// headers From and Date are present, and important ones do not have duplicates.
	// If any of data checkers returns error, it will be reported as DATA
	// command response and command will be considered erroneous.
	DataCheckers []DataChecker

	// DataHandlers are functions to process message body after DATA command.
	// Can be left empty for a NOOP server.
	// If any of DataHandlers returns error, it will be reported as DATA
	// command response and command will be considered erroneous.
	DataHandlers []DataHandler

	// CloseHandlers are called after connection is closed. They can be used to, for example,
	// update counters, save connection metadata into persistent storage like Karma plugin does,
	// or it can even issue shell command to blacklist remote IP by firewall
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
	// Tracer is OpenTelemetry tracer which starts spans for every Transaction
	Tracer trace.Tracer
	// mu guards doneChan and makes closing it and listener atomic from
	// perspective of Serve()
	mu         sync.Mutex
	doneChan   chan struct{}
	listener   *net.Listener
	waitgrp    sync.WaitGroup
	inShutdown atomic.Bool

	// counters for Server.StartPrometheusScrapperEndpoint
	bytesRead                uint64
	bytesWritten             uint64
	transactionsAll          uint64
	transactionsActive       int32
	lastTransactionStartedAt time.Time
}

// startTransaction takes network connection and wraps it into Transaction object to handle all remote
// client interactions via (E)SMTP protocol.
func (srv *Server) startTransaction(c net.Conn) (t *Transaction) {
	var err error
	var ptrs []string
	now := time.Now()
	mu := sync.Mutex{}
	atomic.AddUint64(&srv.transactionsAll, 1)
	atomic.AddInt32(&srv.transactionsActive, 1)
	srv.lastTransactionStartedAt = now
	ctx, cancel := context.WithCancel(context.Background())
	remoteAddr := c.RemoteAddr().(*net.TCPAddr)
	ctxWithTracer, span := srv.Tracer.Start(ctx, "transaction",
		trace.WithSpanKind(trace.SpanKindServer), // важно
		trace.WithAttributes(attribute.String("remote_addr", c.RemoteAddr().String())),
		trace.WithAttributes(attribute.String("remote_ip", remoteAddr.IP.String())),
		trace.WithAttributes(attribute.Int("remote_port", remoteAddr.Port)),
	)
	t = &Transaction{
		ID:        span.SpanContext().TraceID().String(),
		StartedAt: now,

		server:     srv,
		ServerName: srv.Hostname,
		Logger:     srv.Logger,

		Span: span,

		conn:   c,
		reader: bufio.NewReader(srv.wrapWithCounters(c)),
		writer: bufio.NewWriter(srv.wrapWithCounters(c)),
		Addr:   c.RemoteAddr(),
		PTRs:   make([]string, 0),
		ctx:    ctxWithTracer,
		cancel: cancel,

		Aliases:  nil,
		facts:    make(map[string]string, 0),
		counters: make(map[string]float64, 0),
		flags:    make(map[string]bool, 0),
		mu:       &mu,
	}
	t.LogInfo("Starting transaction %s for %s.", t.ID, t.Addr.String())
	// Check if the underlying connection is already TLS.
	// This will happen if the Listener provided Serve()
	// is from tls.Listen()
	var tlsConn *tls.Conn
	tlsConn, t.Encrypted = c.(*tls.Conn)
	if t.Encrypted {
		span.SetAttributes(attribute.Bool("encrypted", true))
		// run handshake otherwise it's done when we first
		// read/write and connection state will be invalid
		err = tlsConn.Handshake()
		if err != nil {
			t.LogError(err, "while performing handshake")
			t.Secured = false
			t.Hate(tlsHandshakeFailedHate)
			span.SetAttributes(attribute.Bool("secured", false))
		} else {
			t.Secured = true
			span.SetAttributes(attribute.Bool("secured", true))
			version, found := TLSVersions[tlsConn.ConnectionState().Version]
			if found {
				t.LogInfo("Connection with %s is already encrypted for server `%s` with %s",
					t.Addr.String(), tlsConn.ConnectionState().ServerName, version,
				)
			} else {
				t.LogWarn("Connection with %s is already encrypted for server `%s` with unknown protocol version %v",
					t.Addr.String(), tlsConn.ConnectionState().ServerName, tlsConn.ConnectionState().Version,
				)
			}
		}
		state := tlsConn.ConnectionState()
		t.TLS = &state
	}
	if !srv.SkipResolvingPTR {
		ptrs, err = t.Resolver().LookupAddr(t.Context(), remoteAddr.IP.String())
		if err != nil {
			t.LogError(err, "while resolving remote address PTR record")
			t.PTRs = make([]string, 0)
		} else {
			t.LogDebug("PTR addresses resolved for %s : %v",
				remoteAddr, ptrs,
			)
			t.PTRs = ptrs
			span.SetAttributes(attribute.StringSlice("ptr", ptrs))
		}
	} else {
		t.LogDebug("PTR resolution disabled")
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
	transaction.mu.Lock()
	defer transaction.mu.Unlock()
	if transaction.closeHandlersCalled {
		transaction.LogDebug("close handlers already called")
		return
	}
	var closeError error
	srv.Logger.Debugf(transaction, "Starting %v close handlers...", len(srv.CloseHandlers))
	for k := range srv.CloseHandlers {
		srv.Logger.Debugf(transaction, "Starting close handler %v...", k)
		closeError = srv.CloseHandlers[k](transaction)
		if closeError != nil {
			transaction.LogError(closeError, "while calling close handler")
		} else {
			transaction.LogDebug("closing handler %v is called", k)
		}
	}
	transaction.closeHandlersCalled = true
	srv.Logger.Infof(transaction, "Closing transaction %s.", transaction.ID)
	atomic.AddInt32(&srv.transactionsActive, -1)
}

// Serve starts the SMTP server and listens on the Listener provided
func (srv *Server) Serve(l net.Listener) error {
	var err error
	var broken bool
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
		for k := range srv.ConnectionCheckers {
			err = srv.ConnectionCheckers[k](transaction)
			if err != nil {
				srv.runCloseHandlers(transaction)
				transaction.LogWarn("%s : after connection checker %v executed", err, k)
				transaction.error(err)
				transaction.close()
				broken = true
				break
			}
		}
		if broken {
			transaction.LogInfo("Connection checkers failed - closing broken transaction...")
			transaction.cancel()
			continue
		}
		if transaction.Encrypted {
			if !transaction.Secured {
				transaction.LogInfo("Connection TLS handshake failed - closing transaction...")
				transaction.cancel()
				continue
			}
		}
		transaction.LogInfo("Accepting connection from %s...", transaction.Addr)

		srv.waitgrp.Add(1)
		go func() {
			defer srv.waitgrp.Done()
			if limiter != nil {
				select {
				case limiter <- struct{}{}:
					transaction.serve()
					<-limiter
				default:
					srv.runCloseHandlers(transaction)
					transaction.reject()
				}
			} else {
				transaction.serve()
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
	if srv.Tracer == nil {
		srv.Tracer = tracesdk.NewTracerProvider().Tracer("msmtpd")
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
