package main

// This example shows how to use msmptd as inbound SMTP server
// using dovecot as persistent storage.
// HELO checkers, Rspamd and open telemetry tracing are enabled.
// We recommend starting Jaeger ui via `docker compose up -d jaeger` and enjoy charts on
// http://127.0.0.1:16686/

import (
	"log"
	"os"
	"path/filepath"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
	"github.com/vodolaz095/msmtpd/plugins/data"
	"github.com/vodolaz095/msmtpd/plugins/dovecot"
	"github.com/vodolaz095/msmtpd/plugins/helo"
	"github.com/vodolaz095/msmtpd/plugins/quarantine"
	"github.com/vodolaz095/msmtpd/plugins/rspamd"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

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
		LtmpSocket:             dovecot.DefaultLMTPSocketPath,
		Timeout:                0,
	}
	// make TLS config never be used for production
	tlsConfig, err := internal.MakeTLSForLocalhost()
	if err != nil {
		log.Fatalf("%s : while making TLS config for localhost", err)
	}
	// how to dial RSPAMD
	rspamdOpts := rspamd.Opts{
		URL: rspamd.DefaultAddress,
		// URL:      "http://holod.local:11334/",
		Password: "who cares",
	}
	// setting up OpenTelemetry
	exp, err := jaeger.New(jaeger.WithAgentEndpoint( // так будет использоваться протокол UDP
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
			semconv.ServiceNameKey.String("dovecot_inbound"),
			attribute.String("environment", "production"),
		)),
	)
	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	// now, we make our server
	server := msmtpd.Server{
		Hostname:         "mx.example.org",
		SkipResolvingPTR: false, // important

		// check HELO/EHLO with sane default values
		HeloCheckers: []msmtpd.HelloChecker{
			// do not check PTR records for clients from local network
			helo.SkipHeloCheckForLocal,
			// hostname should be full top list domain name like mx.mail.ru
			helo.DenyMalformedDomain,
			// you cannot send 127.0.0.1 as HELO/EHLO
			helo.DenyBareIP,
			// you cannot send HELO/EHLO that looks like dynamic addresses issued by ISP to residential internet clients
			helo.DenyDynamicIP,
			// if connection IP address PTR record differs from HELO/EHLO, connection is not allowed
			helo.DenyReverseDNSMismatch,
		},
		RecipientCheckers: []msmtpd.RecipientChecker{
			// check, if recipient is accepted by dovecot by
			// dialing unix:///var/run/dovecot/auth-client
			backend.CheckRecipient,
		},
		DataCheckers: []msmtpd.DataChecker{
			// ensure message has minimal headers in place
			data.CheckHeaders(data.DefaultHeadersToRequire),
			// check message body by rspamd running localy
			// Just point URL to Rspamd fancy webGUI with charts and provide password
			rspamd.CheckByRSPAMD(rspamdOpts),
		},
		DataHandlers: []msmtpd.DataHandler{
			// if rspamd dislikes message, we send it to quarantine directory
			quarantine.MoveToDirectory(filepath.Join(os.TempDir(), "spam")),
			// if rspamd is ok with message, we deliver it to backend
			backend.Deliver,
		},
		TLSConfig: tlsConfig,
		ForceTLS:  true,
		Logger:    &logger,
		Tracer:    tp.Tracer("msmtpd4dovecot"),
	}
	err = server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}
