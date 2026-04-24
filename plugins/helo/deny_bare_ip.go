package helo

import (
	"context"
	"net"

	"github.com/vodolaz095/msmtpd"
	"go.opentelemetry.io/otel/trace"
)

// DenyBareIP denies clients which provide bare IP address in HELO/EHLO command
func DenyBareIP(ctx context.Context, transaction *msmtpd.Transaction) error {
	span := trace.SpanFromContext(ctx)
	if transaction.IsFlagSet(IsLocalAddressFlagName) {
		span.AddEvent("Connection from local address, deny by bare ip is disabled")
		transaction.LogDebug("Connecting from local address %s, DenyBareIP check disabled",
			transaction.Addr.String())
		return nil
	}

	if transaction.IsFlagSet(IsTrustedOrigin) {
		span.AddEvent("Connection from trusted address, deny by bare ip is disabled")
		transaction.LogDebug("Connecting from trusted address %s with accepted helo %s, DenyBareIP check disabled",
			transaction.Addr.String(), transaction.HeloName)
		return nil
	}

	if net.ParseIP(transaction.HeloName) != nil {
		span.AddEvent("HELO/EHLO is bare IP")
		transaction.LogWarn("HELO/EHLO hostname %s is bare ip", transaction.HeloName)
		return complain
	}
	span.AddEvent("HELO/EHLO is not bare IP")
	transaction.LogDebug("HELO/EHLO %s seems to be not bare IP", transaction.HeloName)
	return nil
}
