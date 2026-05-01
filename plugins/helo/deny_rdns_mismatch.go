package helo

import (
	"context"
	"fmt"
	"net"

	"github.com/vodolaz095/msmtpd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DenyReverseDNSMismatch is complicated and very strict test which ensures that
// 1. HELO/EHLO matches any of PTR records resolved for connecting IP
// 2. PTR records of connecting IP are resolved (aka have DNS A records) into IP addresses including connecting IP
// This test prevents delivery from majority small GI domains, which are known to be spammy.
func DenyReverseDNSMismatch(initialCtx context.Context, transaction *msmtpd.Transaction) (err error) {
	raw := transaction.Addr.(*net.TCPAddr).IP
	ctx, span := otel.Tracer("helo.DenyReverseDNSMismatch").Start(initialCtx, "checkHello",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("helo", transaction.HeloName),
			attribute.String("client.addr", raw.String()),
			attribute.StringSlice("client.ptr", transaction.PTRs),
			attribute.Bool("local", transaction.IsFlagSet(IsLocalAddressFlagName)),
			attribute.Bool("trusted", transaction.IsFlagSet(IsTrustedOrigin)),
		))
	defer span.End()

	var found bool
	if transaction.IsFlagSet(IsLocalAddressFlagName) {
		transaction.LogDebug("Connecting from local address %s, DenyReverseDNSMismatch check disabled",
			transaction.Addr.String())
		return nil
	}
	if transaction.IsFlagSet(IsTrustedOrigin) {
		span.AddEvent("Connection from trusted address, RDNS mismatch check is disabled")
		transaction.LogDebug("Connecting from trusted address %s with accepted helo %s, deny by RDNS mismatch check is disabled",
			transaction.Addr.String(), transaction.HeloName)
		return nil
	}

	if len(transaction.PTRs) == 0 {
		span.AddEvent("no ptr records are found")
		transaction.LogWarn("Address %s has no PTR records - dns mismatch detected", transaction.Addr.String())
		return complain
	}
	for i := range transaction.PTRs {
		if transaction.PTRs[i] == transaction.HeloName {
			span.AddEvent("HELO/EHLO is matching RDNS record")
			transaction.LogInfo("HELO/EHLO %s is matching RDNS record %s",
				transaction.HeloName, transaction.PTRs[i])
			found = true
			break
		}
		if transaction.PTRs[i] == transaction.HeloName+"." {
			span.AddEvent("HELO/EHLO is matching RDNS record with dot")
			transaction.LogInfo("HELO/EHLO %s. is matching RDNS record %s",
				transaction.HeloName, transaction.PTRs[i])
			found = true
			break
		}
	}
	if !found {
		span.AddEvent("HELO/EHLO does not match any PTR records")
		transaction.LogWarn("For HELO/EHLO %s there is no matching PTR records", transaction.HeloName)
		return complain
	}
	// now we resolve IP addresses of PTR records
	found = false
	var resolvedPtrs []string
	for i := range transaction.PTRs {
		addrs, errLookup := transaction.Resolver().LookupHost(ctx, transaction.PTRs[i])
		if errLookup != nil {
			span.SetStatus(codes.Error, errLookup.Error())
			span.RecordError(errLookup)
			transaction.LogError(errLookup, fmt.Sprintf("while resolving A record for PTR of %s", transaction.PTRs[i]))
			return complain
		}
		resolvedPtrs = append(resolvedPtrs, addrs...)
		transaction.LogDebug("For PTR %s this addresses resolved %v",
			transaction.PTRs[i], addrs,
		)
	}
	for i := range resolvedPtrs {
		if resolvedPtrs[i] == raw.String() {
			found = true
			span.AddEvent("PTR resolves into connecting address")
			transaction.LogDebug("PTR %s resolves into connecting address %s",
				resolvedPtrs[i], raw.String(),
			)
		}
	}
	if !found {
		span.AddEvent("DNS PTR mismatch!")
		transaction.LogInfo("DNS-PTR mismatch! PTR records %v resolves into %v - remote IP %s is not among them",
			transaction.PTRs, resolvedPtrs, raw.String(),
		)
		return complain
	}
	span.AddEvent("DNS and PTR are correct!")
	transaction.LogInfo("DNS-PTR passes PTR records %v resolves into %v - remote IP %s included!",
		transaction.PTRs, resolvedPtrs, raw.String(),
	)
	return nil
}
