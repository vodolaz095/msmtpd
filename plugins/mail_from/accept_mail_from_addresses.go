package mail_from

import (
	"log"
	"net/mail"

	"msmtpd"
)

// AcceptMailFromAddresses is filter to accept emails only from predefined whilelist of addresses,
// it is more strict version of AcceptMailFromDomains which can accept email from every mailbox of domain
func AcceptMailFromAddresses(whitelist []string) msmtpd.CheckerFunc {
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
	return func(transaction *msmtpd.Transaction) error {
		_, found := goodMailFroms[transaction.MailFrom]
		if found {
			transaction.LogDebug("Sender %s is whitelisted", transaction.MailFrom.String())
			return nil
		}
		return msmtpd.ErrorSMTP{
			Code:    521,
			Message: "I'm sorry, but your email address is not in whitelist",
		}
	}
}
