package msmtpd

import (
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"time"
)

func (t *Transaction) handleDATA(cmd command) {
	var deliverErr error
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
	if err == io.EOF {
		// EOF was reached before MaxMessageSize
		// Accept and deliver message
		t.Body = data.Bytes()
		t.LogDebug("Processing clients message having %v bytes in it", data.Len())
		for k := range t.server.DataHandlers {
			deliverErr = t.server.DataHandlers[k](t)
			if deliverErr != nil {
				t.error(deliverErr)
				return
			}
		}
		t.LogDebug("DATA client message accepted!")
		t.reply(250, "Thank you.")
		t.reset()
	}
	if err != nil {
		t.LogDebug("possible network error: %s", err)
		// Network error, ignore
		return
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
	t.reset()
	return
}
