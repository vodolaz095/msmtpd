package rspamd

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
)

var testRspamdURL, testRspamdPassword string

func TestRspamdEnv(t *testing.T) {
	if os.Getenv("TEST_RSPAMD_URL") == "" {
		t.Errorf("Environment variable TEST_RSPAMD_URL is not set")
	} else {
		testRspamdURL = os.Getenv("TEST_RSPAMD_URL")
	}
	if os.Getenv("TEST_RSPAMD_PASSWORD") == "" {
		t.Errorf("Environment variable TEST_RSPAMD_PASSWORD is not set")
	} else {
		testRspamdPassword = os.Getenv("TEST_RSPAMD_PASSWORD")
	}
}

func TestCheckPyRealRSPAMD(t *testing.T) {
	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      testRspamdURL,
				Password: testRspamdPassword,
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
}

func TestCheckPyMockRSPAMDFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, DefaultAddress+"ping",
		httpmock.NewStringResponder(200, "pong\r\n"))
	// no re
	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      DefaultAddress,
				Password: testRspamdPassword,
				HTTPClient: &http.Client{
					Transport:     httpmock.DefaultTransport,
					CheckRedirect: nil,
					Jar:           nil,
					Timeout:       time.Second,
				},
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() != "421 Too many letters, i cannot read them all now. Please, resend your message later" {
			t.Errorf("%s : wrong status", err)
		}
	} else {
		t.Errorf("greylist not works")
	}
}

func TestCheckPyMockRSPAMDActionNoop(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	rsp, err := httpmock.NewJsonResponder(http.StatusOK, Response{
		Action: ActionNoop,
	})
	if err != nil {
		t.Errorf("%s : while making JSON responder", err)
	}

	httpmock.RegisterResponder(http.MethodPost, DefaultAddress+DefaultEndpoint, rsp)
	httpmock.RegisterResponder(http.MethodGet, DefaultAddress+"ping",
		httpmock.NewStringResponder(200, "pong\r\n"))

	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      DefaultAddress,
				Password: testRspamdPassword,
				HTTPClient: &http.Client{
					Transport:     httpmock.DefaultTransport,
					CheckRedirect: nil,
					Jar:           nil,
					Timeout:       time.Second,
				},
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
}

func TestCheckPyMockRSPAMDActionGreylist(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	rsp, err := httpmock.NewJsonResponder(http.StatusOK, Response{
		Action: ActionGreylist,
	})
	if err != nil {
		t.Errorf("%s : while making JSON responder", err)
	}

	httpmock.RegisterResponder(http.MethodPost, DefaultAddress+DefaultEndpoint, rsp)
	httpmock.RegisterResponder(http.MethodGet, DefaultAddress+"ping",
		httpmock.NewStringResponder(200, "pong\r\n"))

	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      DefaultAddress,
				Password: testRspamdPassword,
				HTTPClient: &http.Client{
					Transport:     httpmock.DefaultTransport,
					CheckRedirect: nil,
					Jar:           nil,
					Timeout:       time.Second,
				},
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() != "451 Your message looks suspicious, try to deliver it one more time, maybe i'll change my mind and accept it" {
			t.Errorf("%s : wrong status", err)
		}
	} else {
		t.Errorf("greylist not works")
	}
}

func TestCheckPyMockRSPAMDActionAddHeader(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	rsp, err := httpmock.NewJsonResponder(http.StatusOK, Response{
		Action: ActionAddHeader,
		Milter: Milter{
			AddHeaders: map[string]AddHeader{
				"Something": {
					Value: "interesting",
					Order: 1,
				},
			}},
	})
	if err != nil {
		t.Errorf("%s : while making JSON responder", err)
	}

	httpmock.RegisterResponder(http.MethodPost, DefaultAddress+DefaultEndpoint, rsp)
	httpmock.RegisterResponder(http.MethodGet, DefaultAddress+"ping",
		httpmock.NewStringResponder(200, "pong\r\n"))

	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      DefaultAddress,
				Password: testRspamdPassword,
				HTTPClient: &http.Client{
					Transport:     httpmock.DefaultTransport,
					CheckRedirect: nil,
					Jar:           nil,
					Timeout:       time.Second,
				},
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
}

func TestCheckPyMockRSPAMDActionSoftReject(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	rsp, err := httpmock.NewJsonResponder(http.StatusOK, Response{
		Action: ActionSoftReject,
	})
	if err != nil {
		t.Errorf("%s : while making JSON responder", err)
	}

	httpmock.RegisterResponder(http.MethodPost, DefaultAddress+DefaultEndpoint, rsp)
	httpmock.RegisterResponder(http.MethodGet, DefaultAddress+"ping",
		httpmock.NewStringResponder(200, "pong\r\n"))

	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      DefaultAddress,
				Password: testRspamdPassword,
				HTTPClient: &http.Client{
					Transport:     httpmock.DefaultTransport,
					CheckRedirect: nil,
					Jar:           nil,
					Timeout:       time.Second,
				},
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() != "421 Too many letters, i cannot read them all now. Please, resend your message later" {
			t.Errorf("%s : wrong status", err)
		}
	} else {
		t.Errorf("greylist not works")
	}
}

func TestCheckPyMockRSPAMDActionHardReject(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	rsp, err := httpmock.NewJsonResponder(http.StatusOK, Response{
		Action: ActionHardReject,
	})
	if err != nil {
		t.Errorf("%s : while making JSON responder", err)
	}

	httpmock.RegisterResponder(http.MethodPost, DefaultAddress+DefaultEndpoint, rsp)
	httpmock.RegisterResponder(http.MethodGet, DefaultAddress+"ping",
		httpmock.NewStringResponder(200, "pong\r\n"))

	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      DefaultAddress,
				Password: testRspamdPassword,
				HTTPClient: &http.Client{
					Transport:     httpmock.DefaultTransport,
					CheckRedirect: nil,
					Jar:           nil,
					Timeout:       time.Second,
				},
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		if err.Error() != "521 Stop sending me this nonsense, please!" {
			t.Errorf("%s : wrong status", err)
		}
	} else {
		t.Errorf("greylist not works")
	}
}

func TestCheckPyMockRSPAMDActionRewriteSubject(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	rsp, err := httpmock.NewJsonResponder(http.StatusOK, Response{
		Action:  ActionRewriteSubject,
		Subject: "new subject rewritten",
	})
	if err != nil {
		t.Errorf("%s : while making JSON responder", err)
	}

	httpmock.RegisterResponder(http.MethodPost, DefaultAddress+DefaultEndpoint, rsp)
	httpmock.RegisterResponder(http.MethodGet, DefaultAddress+"ping",
		httpmock.NewStringResponder(200, "pong\r\n"))

	validMessage := internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net")
	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		DataCheckers: []msmtpd.DataChecker{
			DataChecker(Opts{
				URL:      DefaultAddress,
				Password: testRspamdPassword,
				HTTPClient: &http.Client{
					Transport:     httpmock.DefaultTransport,
					CheckRedirect: nil,
					Jar:           nil,
					Timeout:       time.Second,
				},
			}),
		},
		DataHandlers: []msmtpd.DataHandler{
			func(transaction *msmtpd.Transaction) error {
				for k, v := range transaction.Parsed.Header {
					t.Logf("%s : %v", k, v)
				}
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
	_, err = fmt.Fprintf(wc, validMessage)
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
}
