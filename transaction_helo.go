package msmtpd

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func (t *Transaction) handleHELO(cmd command) {
	ctxWithTracer, span := t.server.Tracer.Start(t.Context(), "handle_helo",
		trace.WithSpanKind(trace.SpanKindInternal), // важно
		trace.WithAttributes(attribute.String("line", cmd.line)),
		trace.WithAttributes(attribute.String("action", cmd.action)),
		trace.WithAttributes(attribute.StringSlice("arguments", cmd.fields)),
		trace.WithAttributes(attribute.StringSlice("params", cmd.params)),
	)
	defer span.End()

	var err error
	if len(cmd.fields) < 2 {
		t.reply(502, "i think you have missed parameter")
		t.Hate(missingParameterPenalty)
		return
	}
	if t.dataHandlersCalledProperly {
		t.LogWarn("HELO called after DATA accepted")
		t.reply(502, "wrong order of commands")
		t.Hate(wrongCommandOrderPenalty)
		return
	}
	if t.HeloName != "" {
		// Reset envelope in case of duplicate HELO
		t.reset()
	}
	t.LogDebug("HELO <%s> is received...", cmd.fields[1])
	t.HeloName = cmd.fields[1]
	t.Protocol = SMTP
	t.Span.SetAttributes(attribute.String("helo", t.HeloName))
	t.Span.SetAttributes(semconv.NetworkProtocolName("smtp"))
	span.SetAttributes(attribute.String("helo", t.HeloName))
	span.SetAttributes(semconv.NetworkProtocolName("smtp"))
	for k := range t.server.HeloCheckers {
		err = t.server.HeloCheckers[k](ctxWithTracer, t)
		if err != nil {
			t.error(err)
			return
		}
	}
	t.LogInfo("HELO <%s> is accepted!", cmd.fields[1])
	t.reply(250, "Go on, i'm listening...")
	t.Love(commandExecutedProperly)
	span.SetStatus(codes.Ok, "accepted")
	return
}

func (t *Transaction) extensions() []string {
	extensions := []string{
		fmt.Sprintf("SIZE %d", t.server.MaxMessageSize),
		"8BITMIME",
		"PIPELINING",
	}
	if t.server.EnableXCLIENT {
		extensions = append(extensions, "XCLIENT")
	}
	if t.server.TLSConfig != nil && !t.Encrypted {
		extensions = append(extensions, "STARTTLS")
	}
	if t.server.Authenticator != nil && t.Encrypted {
		extensions = append(extensions, "AUTH PLAIN LOGIN")
	}
	return extensions
}

func (t *Transaction) handleEHLO(cmd command) {
	ctxWithTracer, span := t.server.Tracer.Start(t.Context(), "handle_ehlo",
		trace.WithSpanKind(trace.SpanKindInternal), // важно
		trace.WithAttributes(attribute.String("line", cmd.line)),
		trace.WithAttributes(attribute.String("action", cmd.action)),
		trace.WithAttributes(attribute.StringSlice("arguments", cmd.fields)),
		trace.WithAttributes(attribute.StringSlice("params", cmd.params)),
	)
	defer span.End()

	var err error
	if len(cmd.fields) < 2 {
		t.reply(502, "i think you have missed parameter")
		t.Hate(missingParameterPenalty)
		return
	}
	if t.dataHandlersCalledProperly {
		t.LogWarn("EHLO called after DATA accepted")
		t.reply(502, "wrong order of commands")
		t.Hate(wrongCommandOrderPenalty)
		return
	}
	if t.HeloName != "" {
		// Reset envelope in case of duplicate HELO
		t.reset()
	}
	t.LogDebug("EHLO <%s> is received...", cmd.fields[1])
	t.HeloName = cmd.fields[1]
	t.Protocol = ESMTP
	t.Span.SetAttributes(attribute.String("ehlo", t.HeloName))
	t.Span.SetAttributes(semconv.NetworkProtocolName("esmtp"))
	span.SetAttributes(attribute.String("ehlo", t.HeloName))
	span.SetAttributes(semconv.NetworkProtocolName("esmtp"))
	for k := range t.server.HeloCheckers {
		err = t.server.HeloCheckers[k](ctxWithTracer, t)
		if err != nil {
			t.error(err)
			return
		}
	}
	t.LogInfo("EHLO <%s> is accepted!", cmd.fields[1])
	fmt.Fprintf(t.writer, "250-%s\r\n", t.server.Hostname)
	extensions := t.extensions()
	if len(extensions) > 1 {
		for _, ext := range extensions[:len(extensions)-1] {
			fmt.Fprintf(t.writer, "250-%s\r\n", ext)
		}
	}
	t.reply(250, extensions[len(extensions)-1])
	t.Love(commandExecutedProperly)
	span.SetStatus(codes.Ok, "accepted")
}
