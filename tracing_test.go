package msmtpd

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"testing"
	"time"

	"github.com/vodolaz095/msmtpd/internal"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const testJaegerHost = "127.0.0.1"

func TestTracingSuccess(t *testing.T) {
	exp, err := jaeger.New(jaeger.WithAgentEndpoint( // так будет использоваться протокол UDP
		jaeger.WithAgentHost("127.0.0.1"),
		jaeger.WithAgentPort("6831"),
	))
	if err != nil {
		t.Errorf("%s : while dialing jaeger", err)
		return
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("msmtpd unit test runner"),
			attribute.String("environment", "unit test"),
		)),
	)
	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("unit-test-success")
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Tracer: tracer,
		CloseHandlers: []CloseHandler{
			func(tr *Transaction) error {
				t.Logf("You can see transaction details on http://%s:16686/trace/%s",
					testJaegerHost, tr.ID,
				)
				return nil
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if supported, _ := c.Extension("AUTH"); supported {
		t.Error("AUTH supported before TLS")
	}
	if supported, _ := c.Extension("8BITMIME"); !supported {
		t.Error("8BITMIME not supported")
	}
	if supported, _ := c.Extension("STARTTLS"); supported {
		t.Error("STARTTLS supported")
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	err = c.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}
	err = c.Verify("foobar@example.net")
	if err == nil {
		t.Error("Unexpected support for VRFY")
	}
	if err = internal.DoCommand(c.Text, 250, "NOOP"); err != nil {
		t.Errorf("NOOP failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
	time.Sleep(time.Second)
	err = tp.Shutdown(context.TODO())
	if err != nil {
		t.Errorf("%s : while flusing traces", err)
	}
}

func TestTracingError(t *testing.T) {
	exp, err := jaeger.New(jaeger.WithAgentEndpoint( // так будет использоваться протокол UDP
		jaeger.WithAgentHost("127.0.0.1"),
		jaeger.WithAgentPort("6831"),
	))
	if err != nil {
		t.Errorf("%s : while dialing jaeger", err)
		return
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("msmtpd unit test runner"),
			attribute.String("environment", "unit test"),
			attribute.String("errors", "present"),
		)),
	)
	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("unit-test-error")
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Tracer: tracer,
		RecipientCheckers: []RecipientChecker{
			func(tr *Transaction, recipient *mail.Address) error {
				tr.LogError(fmt.Errorf("test error"), "it is expected")
				return nil
			},
		},
		CloseHandlers: []CloseHandler{
			func(tr *Transaction) error {
				t.Logf("You can see transaction details on http://%s:16686/trace/%s",
					testJaegerHost, tr.ID,
				)
				return nil
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if supported, _ := c.Extension("AUTH"); supported {
		t.Error("AUTH supported before TLS")
	}
	if supported, _ := c.Extension("8BITMIME"); !supported {
		t.Error("8BITMIME not supported")
	}
	if supported, _ := c.Extension("STARTTLS"); supported {
		t.Error("STARTTLS supported")
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprintf(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	err = c.Reset()
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}
	err = c.Verify("foobar@example.net")
	if err == nil {
		t.Error("Unexpected support for VRFY")
	}
	if err = internal.DoCommand(c.Text, 250, "NOOP"); err != nil {
		t.Errorf("NOOP failed: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
	time.Sleep(time.Second)
	err = tp.Shutdown(context.TODO())
	if err != nil {
		t.Errorf("%s : while flusing traces", err)
	}
}
