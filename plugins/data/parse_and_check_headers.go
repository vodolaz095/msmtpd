package data

import (
	"bytes"
	"net/mail"

	"msmtpd"
)

// Good read - https://en.wikipedia.org/wiki/List_of_SMTP_server_return_codes

const complain = "I cannot parse your message. Do not send me this particular message in future, please, i will never accept it. Thanks in advance!"

// DefaultHeadersToRequire are headers we expect to exist in any incoming email message
var DefaultHeadersToRequire = []string{
	"To",
	"From",
	"Message-ID",
	"Date",
	"Subject",
}

// ParseBodyAndCheckHeaders is Handler for processing message body to ensure is
// 1. parsable as email message
// 2. contains minimal headers required
func ParseBodyAndCheckHeaders(headersRequired []string) func(transaction *msmtpd.Transaction) error {
	return func(transaction *msmtpd.Transaction) error {
		var val string
		message, err := mail.ReadMessage(bytes.NewReader(transaction.Body))
		if err != nil {
			transaction.LogWarn("%s : while parsing message body", err)
			return msmtpd.ErrorSMTP{
				Code:    521,
				Message: complain,
			}
		}
		for _, header := range headersRequired {
			val = message.Header.Get(header)
			if val == "" {
				transaction.LogWarn("header %s is missing", header)
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			transaction.LogDebug("Header %s is %s", header, val)
		}
		transaction.Parsed = message
		transaction.LogInfo("Headers are in place!")
		return nil
	}
}
