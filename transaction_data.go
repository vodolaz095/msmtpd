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

func (t *Transaction) handleDATA(cmd command) {
	var checkErr error
	var deliverErr error

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
				t.Span.RecordError(err)
				t.LogError(checkErr, "while parsing message body")
				t.Hate(tooBigMessagePenalty)
				t.error(ErrorSMTP{
					Code:    521,
					Message: "Stop sending me this nonsense, please!",
				})
				return
			}
			t.LogDebug("Message body of %v bytes is parsed, calling %v DataCheckers on it",
				data.Len(), len(t.server.DataCheckers))
			for j := range t.server.DataCheckers {
				checkErr = t.server.DataCheckers[j](t)
				if checkErr != nil {
					t.error(checkErr)
					t.Span.RecordError(checkErr)
					return
				}
			}
			t.LogInfo("Body (%v bytes) checked by %v DataCheckers successfully",
				data.Len(), len(t.server.DataCheckers))
			t.Love(commandExecutedProperly)

			t.LogDebug("Starting delivery by %v DataHandlers...", len(t.server.DataHandlers))
			for k := range t.server.DataHandlers {
				deliverErr = t.server.DataHandlers[k](t)
				if deliverErr != nil {
					t.error(deliverErr)
					t.Span.RecordError(checkErr)
					return
				}
			}
			t.LogInfo("Message delivered by %v DataHandlers...", len(t.server.DataHandlers))
			t.reply(250, "Thank you.")
			t.Love(commandExecutedProperly)
			t.reset()
			return
		}
		t.Span.RecordError(err)
		t.LogError(err, "possible network error while reading message data")
	}

	// Discard the rest and report an error.
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		t.Span.RecordError(err)
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
