package data

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"msmtpd"
)

// https://rspamd.com/doc/architecture/protocol.html

const RspamdDefaultAddress = "http://localhost:11334/"
const RspamdDefaultEndpoint = "checkv2"

type RspamdOpts struct {
	Url        string
	Password   string
	HttpClient *http.Client
}

type RspamdResponse struct {
	IsSkipped     bool     `json:"is_skipped"`
	Score         float64  `json:"score"`
	RequiredScore float64  `json:"required_score"`
	Action        string   `json:"action"`
	Urls          []string `json:"urls"`
	Emails        []string `json:"emails"`
	MessageId     string   `json:"message-id"`
}

func CheckPyRSPAMD(opts RspamdOpts) func(transaction *msmtpd.Transaction) error {
	if opts.Url == "" {
		opts.Url = RspamdDefaultAddress
	}
	if opts.HttpClient == nil {
		opts.HttpClient = http.DefaultClient
	}

	resp, err := opts.HttpClient.Get(fmt.Sprintf("%sping", opts.Url))
	if err != nil {
		log.Fatalf("%s : while trying to check rspamd server ping on %sping", err, opts.Url)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("wrong status %s while trying to check rspamd server ping on %s", resp.Status, opts.Url)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%s : while reading rspamd server ping response %sping", err, opts.Url)
	}
	pong := string(body)
	if pong != "pong" {
		log.Fatalf("wrong response '%s' while reading rspamd server ping response %sping", pong, opts.Url)
	}
	return func(transaction *msmtpd.Transaction) error {
		payload := bytes.NewReader(transaction.Body)
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", opts.Url, RspamdDefaultEndpoint), payload)
		if err != nil {
			transaction.LogError(err, "error while making HTTP request to RSPAMD")
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: "Error processing message body, please, try againt later",
			}
		}
		req.Header.Add("IP", transaction.Addr.(*net.TCPAddr).IP.String())
		req.Header.Add("Helo", transaction.HeloName)
		req.Header.Add("From", transaction.MailFrom.String())
		//		req.Header.Add("Rcpt", transaction.Rc.String())
		if transaction.Username != "" {
			req.Header.Add("User", transaction.Username)
		}
		if transaction.TLS != nil {
			//req.Header.Add("TLS-Cipher", transaction.TLS.CipherSuite)
			//req.Header.Add("TLS-Cipher", transaction.TLS.Version)
		}
		req = req.WithContext(transaction.Context())
		res, err := opts.HttpClient.Do(req)
		if err != nil {
			transaction.LogError(err, "error while doing HTTP request to RSPAMD")
			return msmtpd.ErrorSMTP{
				Code:    421,
				Message: "Error processing message body, please, try againt later",
			}
		}
		transaction.LogDebug("Response status %s %v", res.Status, res.StatusCode)
		return nil
	}
}
