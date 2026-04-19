package msmtpd

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type command struct {
	line   string
	action string
	fields []string
	params []string
}

func (cmd *command) attachToSpan(span trace.Span) {
	span.SetAttributes(
		attribute.String("cmd.line", cmd.line), attribute.String("cmd.action", cmd.action),
	)
}
