package helo

import (
	"context"
	"strings"

	"github.com/vodolaz095/msmtpd"
	"go.opentelemetry.io/otel/trace"
)

// DenyMalformedDomain checks, if domain in HELO request belongs to top list domains like .ru, .su and so on
func DenyMalformedDomain(ctx context.Context, transaction *msmtpd.Transaction) error {
	span := trace.SpanFromContext(ctx)
	var pass bool
	if transaction.IsFlagSet(IsLocalAddressFlagName) {
		transaction.LogDebug("Connecting from local address %s, deny by malformed domain check disabled",
			transaction.Addr.String())
		return nil
	}
	if transaction.IsFlagSet(IsTrustedOrigin) {
		span.AddEvent("Connection from trusted address, deny by malformed domain check disabled")
		transaction.LogDebug("Connecting from trusted address %s with accepted helo %s, deny by malformed domain check disabled",
			transaction.Addr.String(), transaction.HeloName)
		return nil
	}
	fixed := strings.ToUpper(transaction.HeloName)
	for i := range TopListDomains {
		if pass {
			continue
		}
		if strings.HasSuffix(fixed, "."+TopListDomains[i]) {
			pass = true
		}
		if strings.HasSuffix(fixed, "."+TopListDomains[i]+".") {
			pass = true
		}
	}
	if !pass {
		span.AddEvent("HELO/EHLO is malformed")
		transaction.LogWarn("HELO/EHLO hostname %s is malformed", transaction.HeloName)
		return complain
	}
	span.AddEvent("HELO/EHLO looks ok")
	transaction.LogDebug("HELO/EHLO %s seems to be in top domain list", transaction.HeloName)
	return nil
}
