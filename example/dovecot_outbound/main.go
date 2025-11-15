package main

// This example shows how to use msmptd as outbound SMTP server
// using dovecot as authenticator.
// TLS and open telemetry tracing are enabled.
// We recommend starting Jaeger ui via `docker compose up -d jaeger` and enjoy charts on
// http://127.0.0.1:16686/

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
	"github.com/vodolaz095/msmtpd/plugins/dovecot"
	"github.com/vodolaz095/msmtpd/plugins/procrastinator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const where = "0.0.0.0:1587"

func main() {
	// tune logger
	logger := msmtpd.DefaultLogger{
		Logger: log.Default(),
		Level:  msmtpd.TraceLevel,
	}

	// tune dovecot backend
	backend := dovecot.Dovecot{
		PathToAuthUserDBSocket: dovecot.DefaultAuthUserSocketPath,
		PathToAuthClientSocket: dovecot.DefaultClientSocketPath,
		LmtpSocket:             dovecot.DefaultLMTPSocketPath,
		Timeout:                0,
	}
	// make TLS config never be used for production
	tlsConfig, err := internal.MakeTLSForLocalhost()
	if err != nil {
		log.Fatalf("%s : while making TLS config for localhost", err)
	}
	// add random delays
	worker := procrastinator.Default()

	// setting up OpenTelemetry to report traces to jaeger via udp
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost("127.0.0.1"),
		jaeger.WithAgentPort("6831"),
	))
	if err != nil {
		log.Fatalf("%s : while dialing jaeger", err)
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("dovecot_outbound"),
			attribute.String("environment", "production"),
		)),
	)
	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	// now, we make our server
	server := msmtpd.Server{
		Hostname:         "localhost", // lol
		SkipResolvingPTR: true,        // makes things faster
		TLSConfig:        tlsConfig,
		ForceTLS:         true,
		Logger:           &logger,
		Tracer:           tp.Tracer("msmtpd4dovecot"),

		// perform authentication via unix:///var/run/dovecot/auth-userdb
		Authenticator: backend.Authenticate,

		SenderCheckers: []msmtpd.SenderChecker{
			// ensure user doesn't sends emails from his boss account
			func(_ context.Context, tr *msmtpd.Transaction) error {
				if tr.Username == tr.MailFrom.Address {
					return nil
				}
				tr.LogWarn("User %s tried to send email on behalf of %s",
					tr.Username, tr.MailFrom.Address)
				return msmtpd.ErrorSMTP{
					Code:    535,
					Message: fmt.Sprintf("You are not allowed to send email as different user"),
				}
			},
		},
		DataCheckers: []msmtpd.DataChecker{
			// add random delay to make user believe we send her message somehow
			worker.WaitForData(),
			// ensure message body has valid FROM header matching MAIL FROM
			// to prevent abusing service by sending messages on behalf of somebody else
			func(_ context.Context, tr *msmtpd.Transaction) error {
				froms, fromErr := tr.Parsed.Header.AddressList("From")
				if fromErr != nil {
					tr.LogError(fromErr, "while parsing from header")
					return msmtpd.ErrServiceDoesNotAcceptEmail
				}
				if len(froms) != 1 {
					tr.LogWarn("Duplicate FROM header?")
					return msmtpd.ErrServiceDoesNotAcceptEmail
				}
				if froms[0].Address != tr.MailFrom.Address {
					return msmtpd.ErrorSMTP{
						Code:    535,
						Message: fmt.Sprintf("You are not allowed to send email as different user"),
					}
				}
				return nil
			},
		},
		DataHandlers: []msmtpd.DataHandler{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				// pretend we deliver message
				return nil
			},
		},
	}
	ln, err := tls.Listen("tcp", where, tlsConfig)
	if err != nil {
		log.Fatalf("%s : while starting tls listener on %s", err, where)
	}
	err = server.Serve(ln)
	if err != nil {
		log.Fatalf("%s : while starting server on %s", err, where)
	}
}
