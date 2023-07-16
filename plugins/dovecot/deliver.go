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
	err = expect(pr, "220")
	if err != nil {
		tr.LogError(err, "while getting response to LHLO")
		return temporaryError
	}
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
	for i := range tr.RcptTo {
		err = write(pr, fmt.Sprintf("RCPT TO:<%s>\r\n", tr.RcptTo[i].Address))
		if err != nil {
			tr.LogError(err, "while sending MAIL FROM")
			return temporaryError
		}
		err = expect(pr, "250")
		if err != nil {
			tr.LogError(err, "while getting answer for MAIL FROM")
		} else {
			atLeastOneRecipientFound = true
		}
	}
	if !atLeastOneRecipientFound {
		return permanentError
	}
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
	_, err = pr.W.Write(tr.Body)
	if err != nil {
		tr.LogError(err, "while writing message body")
		return temporaryError
	}
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
