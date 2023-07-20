package data

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

	"msmtpd"
)

var subjectRegex *regexp.Regexp

func init() {
	subjectRegex = regexp.MustCompile(`^Subject:.*$`)
}

// https://rspamd.com/doc/architecture/protocol.html

const RspamdDefaultAddress = "http://localhost:11334/"
const RspamdDefaultEndpoint = "checkv2"

type RspamdOpts struct {
	Url        string
	Password   string
	HttpClient *http.Client
}

type RspamdResponse struct {
	IsSkipped     bool         `json:"is_skipped"`
	Score         float64      `json:"score"`
	RequiredScore float64      `json:"required_score"`
	Action        string       `json:"action"`
	Urls          []string     `json:"urls"`
	Emails        []string     `json:"emails"`
	MessageId     string       `json:"message-id"`
	Subject       string       `json:"subject,omitempty"`
	Milter        RspamdMilter `json:"milter"`
}

type RspamdMilter struct {
	AddHeaders map[string]RspamdAddHeader `json:"add_headers"`
}
type RspamdAddHeader struct {
	Value string
	Order string
}

// RspamdActionNoop is thing rspamd recomends to do with this message
const RspamdActionNoop = "no action"

// RspamdActionGreylist is thing rspamd recomends to do with this message
const RspamdActionGreylist = "greylist"

// RspamdActionAddHeader is thing rspamd recomends to do with this message
const RspamdActionAddHeader = "add header"

// RspamdActionRewriteSubject is thing rspamd recomends to do with this message
const RspamdActionRewriteSubject = "rewrite subject"

// RspamdActionSoftReject is thing rspamd recomends to do with this message
const RspamdActionSoftReject = "soft reject"

// RspamdActionHardReject is thing rspamd recomends to do with this message
const RspamdActionHardReject = "reject"

const rspamdComplain = "Too many letters, i cannot read them all now. Please, resend your message later"

func CheckPyRSPAMD(opts RspamdOpts) func(transaction *msmtpd.Transaction) error {
	if opts.Url == "" {
		opts.Url = RspamdDefaultAddress
	}
	if opts.HttpClient == nil {
		opts.HttpClient = http.DefaultClient
	}
	var setupError error
	resp, setupError := opts.HttpClient.Get(fmt.Sprintf("%sping", opts.Url))
	if setupError != nil {
		log.Fatalf("%s : while trying to check rspamd server ping on %sping", setupError, opts.Url)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("wrong status %s while trying to check rspamd server ping on %s", resp.Status, opts.Url)
	}
	defer resp.Body.Close()
	body, setupError := io.ReadAll(resp.Body)
	if setupError != nil {
		log.Fatalf("%s : while reading rspamd server ping response from %sping", setupError, opts.Url)
	}
	pong := string(body)
	if pong != "pong\r\n" {
		log.Fatalf("wrong response '%s' while reading rspamd server ping response from %sping", pong, opts.Url)
	}
	return func(transaction *msmtpd.Transaction) error {
		payload := bytes.NewReader(transaction.Body)
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", opts.Url, RspamdDefaultEndpoint), payload)
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
			tlsVer, found := msmtpd.TlsVersions[transaction.TLS.Version]
			if found {
				req.Header.Add("TLS-Version", tlsVer)
			}
			req.Header.Add("TLS-Cipher", tls.CipherSuiteName(transaction.TLS.CipherSuite))
		}
		req = req.WithContext(transaction.Context())
		res, err := opts.HttpClient.Do(req)
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
		var rr RspamdResponse
		err = json.Unmarshal(checkResponseBody, &rr)
		if err != nil {
			transaction.LogError(err, "while parsing rspamd response")
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		}
		transaction.LogInfo("Rspamd check result: message `%s` has score %.2f of %.2f required and action is %s",
			rr.MessageId, rr.Score, rr.RequiredScore, rr.Action,
		)
		switch rr.Action {
		case RspamdActionNoop:
			return nil
		case RspamdActionGreylist:
			return msmtpd.ErrorSMTP{
				Code:    451,
				Message: "Your message looks suspicious, try to deliver it one more time, maybe i'll change my mind and accept it",
			}
		case RspamdActionAddHeader:
			for k, v := range rr.Milter.AddHeaders {
				transaction.LogDebug("Rspamd adds header %s : %s", k, v.Value)
				transaction.AddHeader(k, v.Value)
			}
			return nil
		case RspamdActionRewriteSubject:
			transaction.Body = subjectRegex.ReplaceAll(transaction.Body,
				[]byte(fmt.Sprintf("Subject: %s", rr.Subject)),
			)
			return nil
		case RspamdActionSoftReject:
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: rspamdComplain,
			}
		case RspamdActionHardReject:
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
