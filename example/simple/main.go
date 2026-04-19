package main

import (
	"context"
	"log"
	"net/mail"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

func main() {
	logger := msmtpd.DefaultLogger{
		Logger: log.Default(),
		Level:  msmtpd.TraceLevel,
	}

	server := msmtpd.Server{
		Hostname:         "localhost",
		WelcomeMessage:   "Do you believe in our God?",
		MaxConnections:   5,
		MaxMessageSize:   5 * 1024 * 1024, // 5mb
		MaxRecipients:    5,
		SkipResolvingPTR: false, // can make things faster, but various HELO/EHLO checks will not work
		Logger:           &logger,

		// ConnectionCheckers are called when client performed TCP connection
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.LogInfo("Client connects from %s...", tr.Addr.String())
				return nil
			},
		},
		// HeloCheckers are called when client tries to perform HELO/EHLO
		HeloCheckers: []msmtpd.HelloChecker{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.LogInfo("Client HELO is %s", tr.HeloName)
				return nil
			},
			func(_ context.Context, tr *msmtpd.Transaction) error {
				if tr.HeloName != "localhost" {
					tr.Hate(1) // i do not like being irritated
				} else {
					tr.SetFlag("localhost")
				}
				return nil
			},
		},
		// SenderCheckers are called when client provides MAIL FROM to define who sends email message
		SenderCheckers: []msmtpd.SenderChecker{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.LogInfo("Somebody called %s tries to send message...", tr.MailFrom)
				return nil
			},
		},
		// RecipientCheckers are called each time client provides RCPT TO
		// in order to define for whom to send email message
		RecipientCheckers: []msmtpd.RecipientChecker{
			func(_ context.Context, tr *msmtpd.Transaction, recipient *mail.Address) error {
				if strings.HasPrefix(recipient.Address, "info@") {
					return msmtpd.ErrorSMTP{
						Code:    535,
						Message: "Just stop it, please",
					}
				}
				return nil
			},
		},
		// DataCheckers are called on message body to ensure it is properly formatted ham email
		// message according to RFC 5322 and RFC 6532.
		DataCheckers: []msmtpd.DataChecker{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				if tr.Parsed.Header.Get("X-Priority") == "" {
					return msmtpd.ErrorSMTP{
						Code:    535,
						Message: "Please, provide priority for your message!",
					}
				}
				// Add header to message
				tr.AddHeader("Something", "interesting")
				return nil
			},
		},
		// DataHandlers are actual message delivery to persistent storage
		DataHandlers: []msmtpd.DataHandler{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.LogInfo("We pretend we deliver %v bytes of message somehow", len(tr.Body))
				// set float64 fact abount transaction
				tr.Incr("size", float64(len(tr.Body)))
				return msmtpd.ErrServiceNotAvailable
			},
		},
		// CloseHandlers are called when client closes connection, they can be used
		// to, for example, record connection data in database or save metrics
		CloseHandlers: []msmtpd.CloseHandler{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.LogInfo("Closing connection. Karma is %d", tr.Karma())
				// reading string fact
				subject, found := tr.GetFact("subject")
				if found {
					tr.LogInfo("Subject %s", subject)
				}
				// reading float64 counter
				size, found := tr.GetCounter("size")
				if found {
					tr.LogInfo("Body size is %v", size)
				}
				// reading boolean flag present
				if tr.IsFlagSet("localhost") {
					tr.LogInfo("Transaction is send from localhost")
				}
				return nil // error means nothing here, to be honest
			},
		},
	}

	err := server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}
