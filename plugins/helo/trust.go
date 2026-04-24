package helo

import (
	"context"
	"net"

	"github.com/vodolaz095/msmtpd"
	"go.opentelemetry.io/otel/trace"
)

// TrustHellos allows to override all other helo checks for given combination of IP address (as key in map) and HELO greeing (as keys' value)
func TrustHellos(input map[string]string) msmtpd.HelloChecker {
	return func(ctx context.Context, transaction *msmtpd.Transaction) error {
		span := trace.SpanFromContext(ctx)
		a, ok := transaction.Addr.(*net.TCPAddr)
		if !ok {
			// connection via unix socket?
			return nil
		}
		val, found := input[a.IP.String()]
		if found {
			if val == transaction.HeloName {
				span.AddEvent("IP address is found in trusted and HELO match")
				transaction.SetFlag(IsTrustedOrigin)
				return nil
			}
			span.AddEvent("IP address is found in trusted but HELO differs")
			transaction.UnsetFlag(IsTrustedOrigin)
			transaction.LogWarn("IP address %s is found in trusted but HELO differs: expected:%v actual:%s",
				a.IP.String(), val, transaction.HeloName,
			)
			return nil
		}
		span.AddEvent("IP address is not found in trusted")
		transaction.UnsetFlag(IsTrustedOrigin)
		return nil
	}
}
