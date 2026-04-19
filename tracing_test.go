package msmtpd

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"sync"
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
const testJaegerPort = "6831"

func TestTracingSuccess(t *testing.T) {
	exp, err := jaeger.New(jaeger.WithAgentEndpoint( // так будет использоваться протокол UDP
		jaeger.WithAgentHost(testJaegerHost),
		jaeger.WithAgentPort(testJaegerPort),
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
			func(_ context.Context, tr *Transaction) error {
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
	_, err = fmt.Fprint(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
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
		jaeger.WithAgentHost(testJaegerHost),
		jaeger.WithAgentPort(testJaegerPort),
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
			func(_ context.Context, tr *Transaction, recipient *mail.Address) error {
				tr.LogError(fmt.Errorf("test error in RCPT TO"), "it is expected")
				return nil
			},
		},
		CloseHandlers: []CloseHandler{
			func(_ context.Context, tr *Transaction) error {
				tr.LogError(fmt.Errorf("test error in close handler"), "it is expected")

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
	_, err = fmt.Fprint(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
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

func TestTracingConnectionCheckerAndCloseHandlers(t *testing.T) {
	var connectionHandlerCalled, closeHandlerCalled bool
	wg := sync.WaitGroup{}
	wg.Add(2)
	exp, err := jaeger.New(jaeger.WithAgentEndpoint( // так будет использоваться протокол UDP
		jaeger.WithAgentHost(testJaegerHost),
		jaeger.WithAgentPort(testJaegerPort),
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

	tracer := tp.Tracer("unit-test-connection-closer")
	addr, closer := RunTestServerWithoutTLS(t, &Server{
		Tracer: tracer,
		ConnectionCheckers: []ConnectionChecker{
			func(_ context.Context, tr *Transaction) error {
				connectionHandlerCalled = true
				wg.Done()
				return ErrorSMTP{Code: 521, Message: "i do not like you"}
			},
		},
		CloseHandlers: []CloseHandler{
			func(_ context.Context, tr *Transaction) error {
				t.Logf("close handler is called")
				closeHandlerCalled = true
				wg.Done()
				return nil
			},
			func(_ context.Context, tr *Transaction) error {
				t.Logf("You can see transaction details on http://%s:16686/trace/%s",
					testJaegerHost, tr.ID,
				)
				return nil
			},
		},
	})
	defer closer()
	_, err = smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 i do not like you" {
			t.Errorf("%s : wrong error", err)
		}
	} else {
		t.Errorf("Connection not blocked!")
	}
	wg.Wait()
	if !connectionHandlerCalled {
		t.Error("connection handler not called")
	}
	if !closeHandlerCalled {
		t.Error("close handler not called")
	}
	time.Sleep(time.Second)
	err = tp.Shutdown(context.TODO())
	if err != nil {
		t.Errorf("%s : while flusing traces", err)
	}
}
