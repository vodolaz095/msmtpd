package main

// This example shows tracing capabilities
// We recommend starting Jaeger ui via `docker compose up -d jaeger` and enjoy charts on
// http://127.0.0.1:16686/

import (
	"log"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
	"github.com/vodolaz095/msmtpd/plugins/data"
	"github.com/vodolaz095/msmtpd/plugins/helo"
	"github.com/vodolaz095/msmtpd/plugins/sender"

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
	// make TLS config never be used for production
	tlsConfig, err := internal.MakeTLSForLocalhost()
	if err != nil {
		log.Fatalf("%s : while making TLS config for localhost", err)
	}
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
		TLSConfig:        tlsConfig,
		ForceTLS:         true,
		Logger:           &logger,
		Tracer:           tp.Tracer("msmtpd4dovecot"),

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
		SenderCheckers: []msmtpd.SenderChecker{
			// at least require that senders email address belongs to domain we can theoretically
			// deliver messages too. Sane default values restrict 50%+ botnet spam coming
			// from infected routers and smart refrigerators.
			sender.IsResolvable(sender.IsResolvableOptions{
				FallbackToAddressRecord: false,
				AllowLocalAddresses:     false,
			}),
		},
		RecipientCheckers: []msmtpd.RecipientChecker{},
		DataCheckers: []msmtpd.DataChecker{
			data.CheckHeaders(data.DefaultHeadersToRequire),
		},
		DataHandlers: []msmtpd.DataHandler{},
	}
	err = server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}
