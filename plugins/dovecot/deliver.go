package dovecot

import (
	"fmt"

	"msmtpd"
)

// https://en.wikipedia.org/wiki/Local_Mail_Transfer_Protocol
// https://github.com/emersion/go-smtp/blob/master/lmtp_server_test.go
// https://www.rfc-editor.org/rfc/rfc2033.html#section-4.2

func (d *Dovecot) Deliver(tr *msmtpd.Transaction) (err error) {
	pr, err := d.dial("unix", d.LtmpSocket)
	if err != nil {
		tr.LogError(err, "while dialing LMTP socket")
		return temporaryError
	}
	err = expect(pr, "220")
	if err != nil {
		tr.LogError(err, "wrong lmtp greeting")
		return temporaryError
	}

	tr.LogDebug("Sending LHLO localhost into socket %s", d.LtmpSocket)
	err = write(pr, "LHLO localhost\r\n")
	if err != nil {
		tr.LogError(err, "while sending LHLO")
		return temporaryError
	}
	code, features, err := pr.ReadResponse(250)
	if err != nil {
		tr.LogError(err, "while parsing LHLO response")
		return temporaryError
	}
	tr.LogDebug("Response for LHLO: %v %s", code, features)

	tr.LogDebug("Sending MAIL FROM:<%s>", tr.MailFrom.Address)
	err = write(pr, fmt.Sprintf("MAIL FROM:<%s>\r\n", tr.MailFrom.Address))
	if err != nil {
		tr.LogError(err, "while sending MAIL FROM")
		return temporaryError
	}
	err = expect(pr, "250")
	if err != nil {
		tr.LogError(err, "while getting answer for MAIL FROM")
		return temporaryError
	}
	var atLeastOneRecipientFound bool
	if len(tr.Aliases) == 0 {
		for i := range tr.RcptTo {
			tr.LogDebug("Sending RCPT TO:<%s>", tr.RcptTo[i].Address)
			err = write(pr, fmt.Sprintf("RCPT TO:<%s>\r\n", tr.RcptTo[i].Address))
			if err != nil {
				tr.LogError(err, "while sending RCPT TO")
				return temporaryError
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
				return temporaryError
			}
			err = expect(pr, "250")
			if err != nil {
				tr.LogError(err, "while getting answer for RCPT TO")
			} else {
				tr.LogInfo("Alias %s is accepted!", tr.RcptTo[i].String())
				atLeastOneRecipientFound = true
			}
		}
	}
	if !atLeastOneRecipientFound {
		tr.LogError(fmt.Errorf("no recipients allowed"), "no recipients found - both Transaction.RcptTo and Transaction.Aliases are empty")
		return permanentError
	}
	tr.LogDebug("Sending DATA...")
	err = write(pr, "DATA\r\n")
	if err != nil {
		tr.LogError(err, "while sending DATA")
		return temporaryError
	}
	err = expect(pr, "354")
	if err != nil {
		tr.LogError(err, "while getting answer for DATA")
		return temporaryError
	}
	tr.LogDebug("Streaming message...")
	n, err := pr.DotWriter().Write(tr.Body)
	if err != nil {
		tr.LogError(err, "while writing message body")
		return temporaryError
	}
	tr.LogDebug("%v bytes of message is written", n)
	err = pr.DotWriter().Close()
	if err != nil {
		tr.LogError(err, "while writing ending dot")
		return temporaryError
	}
	err = expect(pr, "250")
	if err != nil {
		tr.LogError(err, "while getting answer for message uploaded")
		return temporaryError
	}
	tr.LogDebug("Sending quit")
	err = write(pr, "QUIT")
	if err != nil {
		tr.LogError(err, "while closing connection by QUIT")
		return temporaryError
	}
	tr.LogInfo("Message delivered to dovecot via LMTP %s", d.LtmpSocket)
	return nil
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
