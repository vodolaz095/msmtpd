package rspamd

// Good read
// https://rspamd.com/doc/architecture/protocol.html

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"

	"github.com/vodolaz095/msmtpd"
)

var subjectRegex *regexp.Regexp

func init() {
	subjectRegex = regexp.MustCompile(`^Subject:.*$`)
}

// DefaultAddress is HTTP address where funny RSPAMD GUI is listening
const DefaultAddress = "http://localhost:11334/"

// DefaultEndpoint is endpoint being used for checks
const DefaultEndpoint = "checkv2"

// Opts used to configure how we dial RSPAMD
type Opts struct {
	URL        string
	Password   string
	HTTPClient *http.Client
}

// Response used to parse JSON response of RSPAMD check
type Response struct {
	IsSkipped     bool     `json:"is_skipped"`
	Score         float64  `json:"score"`
	RequiredScore float64  `json:"required_score"`
	Action        string   `json:"action"`
	Urls          []string `json:"urls"`
	Emails        []string `json:"emails"`
	MessageID     string   `json:"message-id"`
	Subject       string   `json:"subject,omitempty"`
	Milter        Milter   `json:"milter"`
}

// Milter is part of Response used to manipulate headers
type Milter struct {
	AddHeaders map[string]AddHeader `json:"add_headers"`
}

// AddHeader is part of Milter in Response used to add headers
type AddHeader struct {
	Value string
	Order string
}

// ActionNoop is thing rspamd recomends to do with this message
const ActionNoop = "no action"

// ActionGreylist is thing rspamd recomends to do with this message
const ActionGreylist = "greylist"

// ActionAddHeader is thing rspamd recomends to do with this message
const ActionAddHeader = "add header"

// ActionRewriteSubject is thing rspamd recomends to do with this message
const ActionRewriteSubject = "rewrite subject"

// ActionSoftReject is thing rspamd recomends to do with this message
const ActionSoftReject = "soft reject"

// ActionHardReject is thing rspamd recomends to do with this message
const ActionHardReject = "reject"

const rspamdComplain = "Too many letters, i cannot read them all now. Please, resend your message later"

// DataChecker is msmtpd.DataChecker function that calls RSPAMD API to validate message against it
func DataChecker(opts Opts) msmtpd.DataChecker {
	if opts.URL == "" {
		opts.URL = DefaultAddress
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}
	var setupError error
	resp, setupError := opts.HTTPClient.Get(fmt.Sprintf("%sping", opts.URL))
	if setupError != nil {
		log.Fatalf("%s : while trying to check rspamd server ping on %sping", setupError, opts.URL)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("wrong status %s while trying to check rspamd server ping on %s", resp.Status, opts.URL)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, setupError := io.ReadAll(resp.Body)
	if setupError != nil {
		log.Fatalf("%s : while reading rspamd server ping response from %sping", setupError, opts.URL)
	}
	pong := string(body)
	if pong != "pong\r\n" {
		log.Fatalf("wrong response '%s' while reading rspamd server ping response from %sping", pong, opts.URL)
	}
	return func(transaction *msmtpd.Transaction) error {
		payload := bytes.NewReader(transaction.Body)
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", opts.URL, DefaultEndpoint), payload)
		if err != nil {
			transaction.LogError(err, "error while making HTTP request to RSPAMD")
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		}
		req.Header.Add("IP", transaction.Addr.(*net.TCPAddr).IP.String())
		req.Header.Add("Helo", transaction.HeloName)
		req.Header.Add("From", transaction.MailFrom.String())
		if opts.Password != "" {
			req.Header.Add("Password", opts.Password)
		}

		for i := range transaction.RcptTo { // Defines SMTP recipient (there may be several Rcpt headers)
			req.Header.Add("Rcpt", transaction.RcptTo[i].String())
		}

		if transaction.Username != "" {
			req.Header.Add("User", transaction.Username)
		}
		if transaction.TLS != nil {
			tlsVer, found := msmtpd.TLSVersions[transaction.TLS.Version]
			if found {
				req.Header.Add("TLS-Version", tlsVer)
			}
			req.Header.Add("TLS-Cipher", tls.CipherSuiteName(transaction.TLS.CipherSuite))
		}
		req = req.WithContext(transaction.Context())
		res, err := opts.HTTPClient.Do(req)
		if err != nil {
			transaction.LogError(err, "error while doing HTTP request to RSPAMD")
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		}
		transaction.LogDebug("Rspamd status %s %v", res.Status, res.StatusCode)
		if res.Body != nil {
			defer res.Body.Close()
		}
		checkResponseBody, err := io.ReadAll(res.Body)
		if err != nil {
			transaction.LogError(err, "error reading Rspamd response")
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		}
		transaction.LogDebug("rspamd response is %s", string(checkResponseBody))
		var rr Response
		err = json.Unmarshal(checkResponseBody, &rr)
		if err != nil {
			transaction.LogError(err, "while parsing rspamd response")
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		}
		transaction.LogInfo("Rspamd check result: message `%s` has score %.2f of %.2f required and action is %s",
			rr.MessageID, rr.Score, rr.RequiredScore, rr.Action,
		)
		switch rr.Action {
		case ActionNoop:
			return nil
		case ActionGreylist:
			return msmtpd.ErrorSMTP{
				Code:    451,
				Message: "Your message looks suspicious, try to deliver it one more time, maybe i'll change my mind and accept it",
			}
		case ActionAddHeader:
			for k, v := range rr.Milter.AddHeaders {
				transaction.LogDebug("Rspamd adds header %s : %s", k, v.Value)
				transaction.AddHeader(k, v.Value)
			}
			return nil
		case ActionRewriteSubject:
			transaction.Body = subjectRegex.ReplaceAll(transaction.Body,
				[]byte(fmt.Sprintf("Subject: %s", rr.Subject)),
			)
			return nil
		case ActionSoftReject:
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		case ActionHardReject:
			return msmtpd.ErrorSMTP{
				Code:    521,
				Message: "Stop sending me this nonsense, please!",
			}
		default:
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		}
	}
}
