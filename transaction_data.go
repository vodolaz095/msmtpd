package msmtpd

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"net/textproto"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// All messages MUST have a 'Date' and 'From' header and a message may not
// contain more than one 'Date', 'From', 'Sender', 'Reply-To', 'To', 'Cc', 'Bcc',
// 'Message-Id', 'In-Reply-To', 'References' or 'Subject' header.
// (c) RFC 5322

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

func (t *Transaction) handleDATA(cmd command) {
	var checkErr error
	var deliverErr error
	var createdAt time.Time
	var from []*mail.Address

	if t.HeloName == "" {
		t.LogDebug("DATA called without HELO/EHLO!")
		t.Hate(missingParameterPenalty)
		t.reply(502, "Please introduce yourself first.")
		return
	}
	if !t.Encrypted && t.server.ForceTLS {
		t.LogDebug("DATA called without STARTTLS!")
		t.Hate(missingParameterPenalty)
		t.reply(502, "Please turn on TLS by issuing a STARTTLS command.")
		return
	}
	if t.server.Authenticator != nil && t.Username == "" {
		t.LogDebug("DATA called without authentication!")
		t.Hate(missingParameterPenalty)
		t.reply(530, "Authentication Required.")
		return
	}
	if t.MailFrom.Address == "" {
		t.Hate(missingParameterPenalty)
		t.LogDebug("DATA called without MAIL FROM!")
		t.reply(502, "It seems you haven't called MAIL FROM in order to explain who sends your message.")
		return
	}
	if len(t.RcptTo) == 0 {
		t.Hate(missingParameterPenalty)
		t.LogDebug("DATA called without RCPT TO!")
		t.reply(502, "It seems you haven't called RCPT TO in order to explain for whom do you want to deliver your message.")
		return
	}
	t.LogDebug("DATA is called...")
	t.reply(354, "Ok, you managed to talk me into accepting your message. Go on, end your data with <CR><LF>.<CR><LF>")
	t.conn.SetDeadline(time.Now().Add(t.server.DataTimeout))
	data := bytes.NewBufferString("")
	reader := textproto.NewReader(t.reader).DotReader()
	_, err := io.CopyN(data, reader, int64(t.server.MaxMessageSize))
	if err != nil {
		if err == io.EOF {
			// EOF was reached before MaxMessageSize, so we can accept and deliver message
			t.Body = data.Bytes()
			t.AddHeader("MSMTPD-Transaction-Id", t.ID)
			t.AddReceivedLine() // will be added as first one
			t.LogDebug("Parsing message body with size %v...", data.Len())
			t.Span.SetAttributes(attribute.Int("size", data.Len()))
			t.Parsed, checkErr = mail.ReadMessage(bytes.NewReader(t.Body))
			if checkErr != nil {
				t.LogWarn("%s : while parsing message body", checkErr)
				t.Hate(tooBigMessagePenalty)
				t.error(ErrorSMTP{
					Code:    521,
					Message: "Stop sending me this nonsense, please!",
				})
				return
			}
			// date header is mandatory according to RFC 5322
			createdAt, checkErr = t.Parsed.Header.Date()
			if checkErr != nil {
				t.LogWarn("%s : while parsing message date", checkErr)
				t.Hate(malformedMessagePenalty)
				t.error(ErrorSMTP{
					Code:    521,
					Message: "Stop sending me this nonsense, please!",
				})
				return
			}
			t.LogInfo("Message created on %s - %s ago",
				createdAt.Format(timeFormatForHeaders),
				time.Since(createdAt).String(),
			)
			// from header is mandatory according to RFC 5322
			from, checkErr = t.Parsed.Header.AddressList("From")
			if checkErr != nil {
				t.LogWarn("%s : while parsing message from header %s",
					checkErr, t.Parsed.Header.Get("From"),
				)
				t.Hate(malformedMessagePenalty)
				t.error(ErrorSMTP{
					Code:    521,
					Message: "Stop sending me this nonsense, please!",
				})
				return
			}
			if len(from) != 1 {
				t.LogWarn("From should contain 1 address")
				t.Hate(malformedMessagePenalty)
				t.error(ErrorSMTP{
					Code:    521,
					Message: "Stop sending me this nonsense, please!",
				})
				return
			}

			// check for duplicate headers
			for _, header := range uniqueHeaders {
				parts, found := t.Parsed.Header[header]
				if found {
					if len(parts) > 1 {
						t.LogWarn("Duplicate header %s %v is found",
							header, parts,
						)
						t.error(ErrorSMTP{
							Code:    521,
							Message: "Stop sending me this nonsense, please!",
						})
					}
				}
			}

			subject := t.Parsed.Header.Get("Subject")
			if subject != "" {
				decoded, decodeErr := decodeBase64EncodedSubject(subject)
				if decodeErr != nil {
					t.LogWarn("%s : while decoding base64 encoded header", decodeErr)
				} else {
					subject = decoded
					t.LogInfo("Subject: %s", subject)
					t.Span.SetAttributes(attribute.String("subject", subject))
					t.SetFact(SubjectFact, subject)
				}
			}

			t.LogDebug("Message body of %v bytes is parsed, calling %v DataCheckers on it",
				data.Len(), len(t.server.DataCheckers))
			for j := range t.server.DataCheckers {
				checkErr = t.server.DataCheckers[j](t)
				if checkErr != nil {
					t.error(checkErr)
					return
				}
			}
			t.LogInfo("Body (%v bytes) checked by %v DataCheckers successfully!",
				data.Len(), len(t.server.DataCheckers))
			t.Love(commandExecutedProperly)

			t.LogDebug("Starting delivery by %v DataHandlers...", len(t.server.DataHandlers))
			for k := range t.server.DataHandlers {
				deliverErr = t.server.DataHandlers[k](t)
				if deliverErr != nil {
					t.error(deliverErr)
					return
				}
			}
			if len(t.server.DataHandlers) > 0 {
				t.LogInfo("Message delivered by %v DataHandlers...", len(t.server.DataHandlers))
			} else {
				t.LogWarn("Message silently discarded - no DataHandlers set...")
			}
			t.reply(250, "Thank you.")
			t.Love(commandExecutedProperly)
			t.reset()
			return
		}
		t.LogError(err, "possible network error while reading message data")
	}

	// Discard the rest and report an error.
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		t.LogDebug("possible network error: %s", err)
		return
	}
	t.reply(552, fmt.Sprintf(
		"Your message is too big, try to say it in less than %d bytes, please!",
		t.server.MaxMessageSize,
	))
	t.Hate(tooBigMessagePenalty)
	t.reset()
	return
}
