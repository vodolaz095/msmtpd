package deliver

import (
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"

	"github.com/vodolaz095/msmtpd"
)

// Good read:
// https://en.wikipedia.org/wiki/Local_Mail_Transfer_Protocol
// https://github.com/emersion/go-smtp/blob/master/lmtp_server_test.go
// https://www.rfc-editor.org/rfc/rfc2033.html#section-4.2

// LMTPOptions allows us to tune how we connect to LMTP server
type LMTPOptions struct {
	// Network can be `unix`, `tcp`, `tcp4` or `tcp6`
	Network string
	// Address shows where we connect
	Address string
	// LHLO is greeting to LMTP server
	LHLO string
	// Timeout limits time of LMTP server interactions
	Timeout time.Duration
}

// String returns where options urges to connect to
func (opts *LMTPOptions) String() string {
	return opts.Network + "//" + opts.Address
}

func dialLMTP(opts LMTPOptions) (*textproto.Conn, error) {
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}
	nc, err := net.DialTimeout(opts.Network, opts.Address, opts.Timeout)
	if err != nil {
		return nil, err
	}
	err = nc.SetDeadline(time.Now().Add(opts.Timeout))
	if err != nil {
		return nil, err
	}
	return textproto.NewConn(nc), nil
}

func expect(conn *textproto.Conn, prefix string) error {
	resp, err := conn.ReadLine()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(resp, prefix) {
		return fmt.Errorf("got %q", resp)
	}
	return nil
}

func write(conn *textproto.Conn, msg string) error {
	_, err := conn.W.Write([]byte(msg))
	if err != nil {
		return err
	}
	return conn.W.Flush()
}

// ViaLocalMailTransferProtocol sends email message via LMTP protocol
func ViaLocalMailTransferProtocol(opts LMTPOptions) msmtpd.DataHandler {
	return func(tr *msmtpd.Transaction) error {
		pr, err := dialLMTP(opts)
		if err != nil {
			tr.LogError(err, "while dialing LMTP socket")
			return TemporaryError
		}
		err = expect(pr, "220")
		if err != nil {
			tr.LogError(err, "wrong LMTP greeting")
			return TemporaryError
		}

		tr.LogDebug("Sending LHLO %s into %s", opts.LHLO, opts.String())
		err = write(pr, "LHLO "+opts.LHLO+"\r\n")
		if err != nil {
			tr.LogError(err, "while sending LHLO")
			return TemporaryError
		}
		code, features, err := pr.ReadResponse(250)
		if err != nil {
			tr.LogError(err, "while parsing LHLO response")
			return TemporaryError
		}
		tr.LogDebug("Response for LHLO: %v %s", code, features)

		tr.LogDebug("Sending MAIL FROM:<%s>", tr.MailFrom.Address)
		err = write(pr, fmt.Sprintf("MAIL FROM:<%s>\r\n", tr.MailFrom.Address))
		if err != nil {
			tr.LogError(err, "while sending MAIL FROM")
			return TemporaryError
		}
		err = expect(pr, "250")
		if err != nil {
			tr.LogError(err, "while getting answer for MAIL FROM")
			return TemporaryError
		}
		var atLeastOneRecipientFound bool
		if len(tr.Aliases) == 0 {
			for i := range tr.RcptTo {
				tr.LogDebug("Sending RCPT TO:<%s>", tr.RcptTo[i].Address)
				err = write(pr, fmt.Sprintf("RCPT TO:<%s>\r\n", tr.RcptTo[i].Address))
				if err != nil {
					tr.LogError(err, "while sending RCPT TO")
					return TemporaryError
				}
				err = expect(pr, "250")
				if err != nil {
					tr.LogError(err, "while getting answer for RCPT TO")
				} else {
					tr.LogInfo("Recipient %s is accepted!", tr.RcptTo[i].String())
					atLeastOneRecipientFound = true
				}
			}
		} else {
			for i := range tr.Aliases {
				tr.LogDebug("Sending RCPT TO:<%s>", tr.Aliases[i].Address)
				err = write(pr, fmt.Sprintf("RCPT TO:<%s>\r\n", tr.Aliases[i].Address))
				if err != nil {
					tr.LogError(err, "while sending RCPT TO")
					return TemporaryError
				}
				err = expect(pr, "250")
				if err != nil {
					tr.LogError(err, "while getting answer for RCPT TO")
				} else {
					tr.LogInfo("Alias %s is accepted!", tr.Aliases[i].String())
					atLeastOneRecipientFound = true
				}
			}
		}
		if !atLeastOneRecipientFound {
			tr.LogError(fmt.Errorf("no recipients allowed"), "no recipients found - both Transaction.RcptTo and Transaction.Aliases are empty")
			return UnknownRecipientError
		}
		tr.LogDebug("Sending DATA...")
		err = write(pr, "DATA\r\n")
		if err != nil {
			tr.LogError(err, "while sending DATA")
			return TemporaryError
		}
		err = expect(pr, "354")
		if err != nil {
			tr.LogError(err, "while getting answer for DATA")
			return TemporaryError
		}
		tr.LogDebug("Streaming message...")
		n, err := pr.DotWriter().Write(tr.Body)
		if err != nil {
			tr.LogError(err, "while writing message body")
			return TemporaryError
		}
		tr.LogDebug("%v bytes of message is written", n)
		err = pr.DotWriter().Close()
		if err != nil {
			tr.LogError(err, "while writing ending dot")
			return TemporaryError
		}
		err = expect(pr, "250")
		if err != nil {
			tr.LogError(err, "while getting answer for message uploaded")
			return TemporaryError
		}
		tr.LogDebug("Sending quit")
		err = write(pr, "QUIT")
		if err != nil {
			tr.LogError(err, "while closing connection by QUIT")
			return TemporaryError
		}
		tr.LogInfo("Message delivered to dovecot via LMTP %s", opts.String())
		return nil
	}
}

/*

[vodolaz095@holod ~]$ swaks --protocol lmtp --lhlo localhost --socket /run/dovecot/lmtp --to=vodolaz095@localhost --from=anatolij@vodolaz095.ru
=== Trying /run/dovecot/lmtp...
=== Connected to /run/dovecot/lmtp.
<-  220 localhost Dovecot ready.
 -> LHLO localhost
<-  250-localhost
<-  250-8BITMIME
<-  250-CHUNKING
<-  250-ENHANCEDSTATUSCODES
<-  250-PIPELINING
<-  250 STARTTLS
 -> MAIL FROM:<anatolij@vodolaz095.ru>
<-  250 2.1.0 OK
 -> RCPT TO:<vodolaz095@localhost>
<-  250 2.1.5 OK
 -> DATA
<-  354 OK
 -> Date: Thu, 20 Jul 2023 09:35:00 +0300
 -> To: vodolaz095@localhost
 -> From: anatolij@vodolaz095.ru
 -> Subject: test Thu, 20 Jul 2023 09:35:00 +0300
 -> Message-Id: <20230720093500.1700966@localhost>
 -> X-Mailer: swaks v20181104.0 jetmore.org/john/code/swaks/
 ->
 -> This is a test mailing
 ->
 ->
 -> .
<-  250 2.0.0 <vodolaz095@localhost> s8G2G5TVuGRV8xkA0J78UA Saved
 -> QUIT
<-  221 2.0.0 Bye

*/
