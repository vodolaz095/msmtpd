package data

import (
	"time"

	"github.com/vodolaz095/msmtpd"
)

// Good read - https://en.wikipedia.org/wiki/List_of_SMTP_server_return_codes

// All messages MUST have a 'Date' and 'From' header and a message may not
// contain more than one 'Date', 'From', 'Sender', 'Reply-To', 'To', 'Cc', 'Bcc',
// 'Message-Id', 'In-Reply-To', 'References' or 'Subject' header.
// (c) RFC 5322

const complain = "I cannot parse your message. Do not send me this particular message in future, please, i will never accept it. Thanks in advance!"

// DefaultHeadersToRequire are headers we expect to exist in any incoming email message
var DefaultHeadersToRequire = []string{
	"Date",
	"From",

	"To",
	"Message-ID",
	"Subject",
}

// mandatoryHeaders are headers should always be present in email message according to RFC 5322
var mandatoryHeaders = []string{
	"Date",
	"From",
}

// uniqueHeaders are headers that should not have duplicates according to RFC 5322
var uniqueHeaders = []string{
	"Date",
	"From",
	"Sender",
	"Reply-To",
	"To",
	"Cc",
	"Bcc",
	"Message-Id",
	"In-Reply-To",
	"References",
	"Subject",
}

const tooOld = 15 * 24 * time.Hour
const tooFarInFuture = 2 * 24 * time.Hour

// CheckHeaders is DataChecker for processing message body to ensure it is compatible with RFC 5322 and
// contains minimal headers required. This checker protects from the majority of malformed messages
// that can break, for example, antique Outlook Express and probably other email clients.
func CheckHeaders(headersRequired []string) msmtpd.DataChecker {
	return func(transaction *msmtpd.Transaction) error {
		var val string
		for _, header := range mandatoryHeaders {
			val = transaction.Parsed.Header.Get(header)
			if val == "" {
				transaction.LogWarn("mandatory header %s is missing", header)
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			transaction.LogDebug("Header %s is %s", header, val)
		}
		for _, header := range headersRequired {
			val = transaction.Parsed.Header.Get(header)
			if val == "" {
				transaction.LogWarn("required header %s is missing", header)
				return msmtpd.ErrorSMTP{
					Code:    521,
					Message: complain,
				}
			}
			transaction.LogDebug("Header %s is %s", header, val)
		}
		for _, header := range uniqueHeaders {
			parts, found := transaction.Parsed.Header[header]
			if found {
				if len(parts) > 1 {
					transaction.LogWarn("Duplicate header %s %v is found",
						header, parts,
					)
					return msmtpd.ErrorSMTP{
						Code:    521,
						Message: complain,
					}
				}
			}
		}
		timestamp, err := transaction.Parsed.Header.Date()
		if err != nil {
			transaction.LogWarn("Malformed date header with value %s",
				transaction.Parsed.Header.Get("Date"))
			return msmtpd.ErrorSMTP{
				Code:    521,
				Message: complain,
			}
		}
		transaction.LogInfo("Message was generated on %s", timestamp.Format(time.ANSIC))
		if time.Since(timestamp) > tooOld {
			transaction.LogWarn("Message is too old: %s", time.Since(timestamp).String())
			return msmtpd.ErrorSMTP{
				Code:    521,
				Message: complain,
			}
		}
		if time.Now().Add(tooFarInFuture).Before(timestamp) {
			transaction.LogWarn("Message is too far away in future")
			return msmtpd.ErrorSMTP{
				Code:    521,
				Message: complain,
			}
		}
		transaction.LogInfo("Headers are in place!")
		return nil
	}
}
