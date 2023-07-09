package mail_from

import (
	"log"
	"net/mail"

	"msmtpd"
)

func AcceptMailFromAddresses(whitelist []string) msmptd.CheckerFunc {
	var err error
	var parsed *mail.Address
	goodMailFroms := make(map[mail.Address]bool, 0)

	for i, raw := range whitelist {
		parsed, err = mail.ParseAddress(raw)
		if err != nil {
			log.Fatalf("%s : while plugin mail_from/AcceptMailFromAddresses tries to parse address %v %s",
				err, i, raw,
			)
		}
		goodMailFroms[*parsed] = true
	}
	return func(transaction *msmptd.Transaction, name string) error {
		_, found := goodMailFroms[transaction.MailFrom]
		if found {
			transaction.LogDebug("Sender %s is whitelisted", transaction.MailFrom.String())
			transaction.SetFact(MailFromWhileListed, "true")
			return nil
		}
		return msmptd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but your email address is not in whitelist",
		}
	}
}
