package msmtpd

import (
	"crypto/tls"
	"net/smtp"
	"testing"

	"msmtpd/internal"
)

func TestAuthRejection(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysFails,
		ForceTLS:      true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Auth(smtp.PlainAuth("foo", "foo", "bar", "127.0.0.1")); err == nil {
		t.Error("Auth worked despite rejection")
	}
}

func TestAuthNotSupported(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
		ForceTLS: true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Auth(smtp.PlainAuth("foo", "foo", "bar", "127.0.0.1")); err == nil {
		t.Error("Auth worked despite no authenticator")
	}
}

func TestAuthBypass(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysFails,
		ForceTLS:      true,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Error("Unexpected MAIL success")
	}
}

func TestAuthPlain(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
		ForceTLS:      true,
		Authenticator: AuthenticatorForTestsThatAlwaysWorks,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	err = c.Auth(smtp.PlainAuth("", "mr.bubbles", "who cares", "127.0.0.1"))
	if err != nil {
		t.Errorf("%s : while performing plain authorization", err)
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while performing plain authorization", err)
	}
}

func TestLOGINAuth(t *testing.T) {
	addr, closer := RunTestServerWithTLS(t, &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysWorks,
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 334, "AUTH LOGIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "foo"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = internal.DoCommand(c.Text, 334, "AUTH LOGIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = internal.DoCommand(c.Text, 334, "Zm9v"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "foo"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = internal.DoCommand(c.Text, 334, "AUTH LOGIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = internal.DoCommand(c.Text, 334, "Zm9v"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = internal.DoCommand(c.Text, 235, "Zm9v"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}

func TestAuthenticationErrors(t *testing.T) {
	cert, err := tls.X509KeyPair(internal.LocalhostCert, internal.LocalhostKey)
	if err != nil {
		t.Errorf("Cert load failed: %v", err)
	}
	server := &Server{
		Authenticator: AuthenticatorForTestsThatAlwaysWorks,
	}
	addr, closer := RunTestServerWithoutTLS(t, server)
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		t.Errorf("Dial failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "AUTH PLAIN foobar"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = c.Hello("localhost"); err != nil {
		t.Errorf("HELO failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "AUTH PLAIN foobar"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Errorf("MAIL didn't fail")
	}
	if err = internal.DoCommand(c.Text, 502, "STARTTLS"); err != nil {
		t.Errorf("STARTTLS didn't fail: %v", err)
	}
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Errorf("STARTTLS failed: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "AUTH UNKNOWN"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "AUTH PLAIN foobar"); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = internal.DoCommand(c.Text, 502, "AUTH PLAIN Zm9vAGJhcg=="); err != nil {
		t.Errorf("AUTH didn't fail: %v", err)
	}
	if err = internal.DoCommand(c.Text, 334, "AUTH PLAIN"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = internal.DoCommand(c.Text, 235, "Zm9vAGJhcgBxdXV4"); err != nil {
		t.Errorf("AUTH didn't work: %v", err)
	}
	if err = c.Mail("sender@example.org"); err != nil {
		t.Errorf("MAIL failed: %v", err)
	}
	if err = c.Mail("sender@example.org"); err == nil {
		t.Errorf("Duplicate MAIL didn't fail")
	}
	if err = c.Quit(); err != nil {
		t.Errorf("Quit failed: %v", err)
	}
}
