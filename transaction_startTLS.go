package msmtpd

import (
	"bufio"
	"crypto/tls"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (t *Transaction) handleSTARTTLS(cmd command) {
	_, span := t.server.Tracer.Start(t.Context(), "handle_start_tls",
		trace.WithSpanKind(trace.SpanKindInternal), // важно
		trace.WithAttributes(attribute.String("line", cmd.line)),
		trace.WithAttributes(attribute.String("action", cmd.action)),
		trace.WithAttributes(attribute.StringSlice("arguments", cmd.fields)),
		trace.WithAttributes(attribute.StringSlice("params", cmd.params)),
	)
	defer span.End()
	var err error
	if t.Encrypted {
		t.LogDebug("Connection is already encrypted!")
		t.reply(502, "Already running in TLS")
		return
	}
	if t.server.TLSConfig == nil {
		t.reply(502, "TLS not supported")
		return
	}
	t.LogDebug("STARTTLS [%s] is received...", cmd.line)
	tlsConn := tls.Server(t.conn, t.server.TLSConfig)
	t.reply(220, "Connection is encrypted, we can talk freely now!")
	err = tlsConn.Handshake()
	if err != nil {
		t.LogError(err, "couldn't perform handshake")
		t.reply(550, "TLS Handshake error")
		return
	}
	t.LogInfo("Connection is encrypted via StartTLS!")
	t.Span.SetAttributes(attribute.Bool("encrypted", true))
	// Reset envelope as a new EHLO/HELO is required after STARTTLS
	t.reset()
	// Reset deadlines on the underlying connection before I replace it
	// with a TLS connection
	err = t.conn.SetDeadline(time.Time{})
	if err != nil {
		t.LogError(err, "error setting deadline for encrypted connection")
		t.reply(550, "TLS Handshake error")
		return
	}

	// Replace connection with a TLS connection
	t.conn = tlsConn
	t.reader = bufio.NewReader(tlsConn)
	t.writer = bufio.NewWriter(tlsConn)
	t.scanner = bufio.NewScanner(t.reader)
	t.Encrypted = true
	// Save connection state on peer
	state := tlsConn.ConnectionState()
	t.TLS = &state
	// Flush the connection to set new timeout deadlines
	t.flush()
	t.Love(commandExecutedProperly)
	span.SetStatus(codes.Ok, "connection is encrypted")
}
