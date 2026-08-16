# Agent Guide for msmtpd

msmtpd is a **Go-based SMTP server** that supports plugins, tracing, metrics, and transparent proxying. This guide covers essential context for agents working in this repository.

## Overview

- **Core behavior**: Implements the SMTP/EHLO protocol; validates sender/recipient/message; forwards via SMTP/Dovecot/LMTP; supports auth (STARTTLS+PLAIN/LOGIN), TLS (force/enforce options), XCLIENT, Proxy protocol.
- **Plugin architecture**: Checkers (ConnectionChecker, HelloChecker, SenderChecker, RecipientChecker, DataChecker, AuthenticatorFunc), Handlers (DataHandler), CloseHandlers. Each plugin implements callback functions into the server lifecycle. Plugins can be added as slices to the `Server` struct or via config. Plugin code lives under `plugins/`.
- **Server loop**: `Serve()` calls `startTransaction()` per connection. Each connection runs `ConnectionCheckers`; then TLS handshake (if applicable); then background goroutine with `transaction.serve()` handles one SMTP transaction.

## Essential Commands

**Build / Lint**
```bash
# Full build and test
go test -v ./...

# Run only msmtpd package tests
go test -v ./...

# Lint / type check
gofmt -w=true -s=true -l=true ./
golint ./...
go vet ./...

# Dependencies and vuln checks
go mod tidy
go mod verify
govulncheck ./...  # via Makefile: `make vuln`
```

**Tests**
```bash
# Core msmtpd tests (includes server.go, transaction*.go)
go test -v ./...

# Check transaction data parsing
go test -v ./transaction_test.go

# Transaction data tests (critical for message parsing)
go test -v ./transaction_data_test.go  # ~4200 lines
```

**Run / Development**
```bash
# Minimal development server (non-production, uses example/minimal/main.go)
go run example/minimal/main.go

# Simple development server
go run example/simple/main.go

# Full msmtpd server (plugins + tracing + metrics)
go run internal/main.go  # check internal/ directory for main entry

# Test against server via swaks client
swaks --to recipient1@example.org,recipient2@example.org,recipient3@example.org,recipient4@example.org \
  --from sender@example.org \
  --server localhost --port 1025 \
  --timeout 600
  # Check test command in Makefile
```

**Deploy / Docker**
- Run via `docker-compose.yaml`: uses `jaeger:1.47.0` backend, `redis:7-alpine` for metrics storage.
- Prometheus endpoints exposed via `StartPrometheusScrapperEndpoint()`.

## Architecture and Data Flow

### SMTP Session Flow (E)SMTP protocol
1. TCP connect → `ConnectionCheckers` (reject on error)
2. HELO/EHLO → `HeloCheckers`
3. MAIL FROM → `SenderCheckers`
4. RCPT TO → `RecipientCheckers` (per-recipient validation)
5. Data commands → `DataCheckers` then `DataHandlers` (forwards message)
6. Session end → `CloseHandlers` (karma, counters)

### Plugin Callbacks (Types in `server.go` ~120 lines)
```go
type ConnectionChecker CheckerFunc              // called on TCP connect
type HelloChecker CheckerFunc                   // after HELO/EHLO
type SenderChecker CheckerFunc                  // after MAIL FROM
type RecipientChecker func(ctx, t, r *mail.Address) error  // per RCPT TO
type DataChecker CheckerFunc                    // before DATA
type DataHandler CheckerFunc                    // after DATA (forwarding)
type CloseHandler CheckerFunc                   // on close
type AuthenticatorFunc func(ctx, t, user, pass string) error  // STARTTLS + auth
```

### Transaction Lifecycle
- Created in `startTransaction()` (main flow in `server.go`).
- Fields: `reader`, `writer` (wraps `bufio.Scanner` + counters); `Span` (OpenTelemetry span); `Tracer` (OpenTelemetry); `Logger`.
- Uses `ctx: context.WithCancel` for transaction cancellation on disconnect/abort.
- Counters (`bytesRead`, `bytesWritten`, `transactionsSuccess/Fail`, `transactionsActive`) track metrics.

### Plugins
- **Connection**: blacklist/whitelist, RBL checks (RBL plugin).
- **Sender**: sender reputation (senderscore plugin).
- **Recipient**: BCC list (recipient plugin).
- **Data**: message validation (data plugin, rspamd plugin).
- **Deliver**: Dovecot LMTP integration, SMTP relay (dovecot, deliver plugins).
- **Quarantine**: quarantine plugin.
- **Karma**: stores per-IP statistics and "karma" for scoring (karma plugin).
- **Sender**: sender plugin.

### Key Modules
- `transaction*.go` (~15 files, ~25000 lines): core transaction implementation, message parsing, headers, AUTH, HELO, MAIL, RCPT, DATA, HELLO commands.
- `server.go` (~20000 lines): main server loop, plugin management, OpenTelemetry integration, Prometheus metrics.
- `helpers.go` (~20 lines): utility functions (line wrapping, base64 subject decoding).
- `logger.go` (~20000 lines): logger interface used by all components and plugins.
- `transaction_auth.go` (~20000 lines): auth handling (PLAIN/LOGIN).
- `transaction_data.go` (~20000 lines): DATA command parsing and message validation.

### Counters and Facts
- **Counters**: `tx.counters["foo"] = 1.2` (stored in transaction `counters` map, used by Karma plugins).
- **Facts**: `tx.facts["foo"] = "bar"` (stored in transaction `facts` map).
- **Karma plugin**: uses counters, `tlsHandshakeFailedHate`, `wrongCommandOrderPenalty`, etc. (~40 lines of penalty constants).
- **Subject fact**: `SubjectFact = "subject"`.
- **Null sender flag**: `NullSenderFlag = "null_sender"`.

### OpenTelemetry Jaeger Tracing
- Uses `go.opentelemetry.io/otel` package, spans every `Transaction`.
- Spans have `trace.SpanKindServer` and attributes: OS, architecture, server address, client IP/port, `ptr` array, `encrypted` bool.
- Jaeger backend: `http://localhost:16686` UI, UDP 6831 for spans.
- See `tracing_test.go` for usage patterns.

## Code Style / Conventions

- **Go 1.26** (see `go.mod`).
- **Import path**: `github.com/vodolaz095/msmtpd`.
- **Package name**: `package msmtpd` (main), each plugin uses its own package name (e.g., `connection`), transaction/transaction*.go are `package msmtpd`.
- **Logging**: uses `Logger` interface, `LogInfo`, `LogWarn`, `LogError`, `LogDebug`.
- **Error handling**: returns errors to transaction for logging via `transaction.LogError(err, reason)` or `transaction.error(err)`.
- **Atomic flags**: many state uses atomic operations (`atomic.AddInt32`, `atomic.AddUint64`, `atomic.Bool`), transaction lock `mu` (`sync.Mutex`).
- **Transaction cancellation**: `transaction.ctx` is `context.WithCancel` context; client disconnect/timeout calls cancel.

## Known Non-Obvious Patterns

### Plugin Registration
Plugins must register their callbacks by assigning them to fields in the `Server` struct, e.g.:
```go
server.CloseHandlers = append(server.CloseHandlers, myCloseHandler)
```

### Transaction Context
Every transaction has its own `ctx` via `context.WithCancel(srv.Context)`. Use this for cancellation when connection closes or error occurs. Do not share the server's context within a transaction unless you want the same cancellation behavior for all transactions.

### Counters Usage
- Counters are **transaction-local** (stored in `t.counters`).
- Use `add` operation `tx.addCounter("foo", 1.0)` which returns `tx`.
- Karma plugin uses per-IP counters: `tx.addCounter(ip, transaction.counters).Value(ipString)`.

### Line Wrapping (RFC 5322 compliance)
`helpers.go`: `wrap()` handles SMTP line length limit (76 chars).

## Common Gotchas

- **Line wrapping**: `wrap()` function ensures lines ≤76 characters in SMTP messages.
- **XCLIENT and Proxy enabled by default**: both are **disabled** by default (`EnableXCLIENT`, `EnableProxyProtocol`).
- **MaxRecipients**: `MaxRecipients` defaults to **100**.
- **ForceTLS**: requires `TLSConfig` to be non-`nil`.
- **Transaction cancellation**: cancel context when closing or error.
- **Base64 subject encoding**: `decodeBase64EncodedSubject()` handles `=?UTF-8?B?` and `=?utf-8?b?` syntax.
- **OpenTelemetry tracing**: span is started in `startTransaction()`.

## Project Structure

```
msmtpd/
├── server.go           # server loop, plugins, metrics (~20000 lines)
├── transaction*.go     # transaction handling, message parsing (many files, ~25000 lines)
├── helpers.go          # utilities (20000 lines)
├── logger.go           # logging interface (20000 lines)
├── plugins/            # plugin directory
│   ├── connection/     # blacklist, RBL, senderscore
│   ├── dovecot/        # LMTP/Dovecot integration
│   ├── deliver/        # SMTP relay
│   ├── data/           # message validation
│   ├── helo/           # hostname checks
│   ├── karma/          # per-IP statistics
│   ├── recipient/      # BCC list
│   ├── rspamd/         # spam filtering
│   ├── sender/         # sender checks
├── internal/           # internal tools
└── example/            # example servers (minimal, simple, smtp_proxy, etc.)
```

## Key Files Reference

- `server.go` (~20000 lines): server loop, plugins, metrics
- `transaction*.go` (~25000 lines): transaction handling, header parsing
- `transaction_data.go` (~7000 lines): DATA command and message parsing
- `transaction_test.go` (~15000 lines): transaction integration tests
