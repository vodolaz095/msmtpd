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
	err = write(pr, "LHLO localhost\r\n")
	if err != nil {
		tr.LogError(err, "while sending LHLO")
		return temporaryError
	}
	err = expect(pr, "250")
	if err != nil {
		tr.LogError(err, "while getting response to LHLO")
		return temporaryError
	}
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
				atLeastOneRecipientFound = true
			}
		}
	}
	if !atLeastOneRecipientFound {
		tr.LogError(fmt.Errorf("NO_RECEPIENTS"), "no recipients found - both Transaction.RcptTo and Transaction.Aliases are empty")
		return permanentError
	}
	tr.LogDebug("Sending DATA")
	err = write(pr, "DATA\r\n")
	if err != nil {
		tr.LogError(err, "while sending DATA")
		return temporaryError
	}
	err = expect(pr, "354")
	if err != nil {
		tr.LogError(err, "while getting answer for MAIL FROM")
		return temporaryError
	}
	n, err := pr.W.Write(tr.Body)
	if err != nil {
		tr.LogError(err, "while writing message body")
		return temporaryError
	}
	tr.LogDebug("%v bytes of message is written", n)
	_, err = pr.W.WriteString("\r\n.\r\n")
	if err != nil {
		tr.LogError(err, "while writing ending dot")
		return temporaryError
	}
	err = expect(pr, "250")
	if err != nil {
		tr.LogError(err, "while getting answer for MAIL FROM")
		return temporaryError
	}
	err = write(pr, "QUIT")
	if err != nil {
		tr.LogError(err, "while closing connection by QUIT")
		return temporaryError
	}
	return nil
}
