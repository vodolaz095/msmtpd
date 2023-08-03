package main

// This is simple SMTP proxy
// it accepts emails with defined range of senders and recipients, ant
// than it delivers messages using 3rd party SMTP server

import (
	"crypto/tls"
	_ "embed"
	"log"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/plugins/data"
	"github.com/vodolaz095/msmtpd/plugins/deliver"
	"github.com/vodolaz095/msmtpd/plugins/recipient"
	"github.com/vodolaz095/msmtpd/plugins/sender"
)

//go:embed key.pem
var key []byte

//go:embed cert.pem
var cert []byte

func main() {
	// it can be a good idea to implement some config handling
	// but i do not want to make example overcomplicated

	logger := msmtpd.DefaultLogger{
		Logger: log.Default(),
		Level:  msmtpd.InfoLevel,
	}
	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		log.Fatalf("%s : while setting certificate", err)
	}
	// we configure 3rd party SMTP proxy (tested with smtp.yandex.ru, haraka, postfix) that performs
	// actual delivery of email messages
	proxyOptions := deliver.SMTPProxyOptions{
		Network:  "tcp",
		Address:  "smtp.example.org:25",
		HELO:     "localhost",
		TLS:      &tls.Config{},
		Auth:     smtp.PlainAuth("", "username", "password", "smtp.example.org"),
		MailFrom: "", // pass as is from incoming transaction, if not null, we can override it here
		// pass as is from incoming Transaction.Aliases (if present), or Transaction.RcptTo,
		// if not null, we can set override it here, so, no matter who is recipient, message is send to
		// overrides
		RcptTo: nil,
	}

	server := msmtpd.Server{
		Hostname:         "mx.example.org",
		WelcomeMessage:   "I want to drink tea and eat meat!",
		MaxConnections:   5,
		MaxMessageSize:   5 * 1024 * 1024, // 5mb
		MaxRecipients:    5,
		SkipResolvingPTR: false, // can make things faster, but various HELO/EHLO checks will not work

		ForceTLS: true,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certificate},
		},

		// SenderCheckers are called when client provides MAIL FROM to define who sends email message
		SenderCheckers: []msmtpd.SenderChecker{
			func(tr *msmtpd.Transaction) error {
				tr.LogInfo("Somebody called %s tries to send message...", tr.MailFrom.String())
				return nil
			},
			sender.AcceptMailFromDomainsOrAddresses([]string{
				// we accept all emails from yandex.ru and mail.ru
				"yandex.ru",
				"mail.ru",
			}, []string{
				"anatolij@vodolaz095.ru", // we accept all send by Anatolij (but it can have consequences)
			}),
		},
		// RecipientCheckers are called each time client provides RCPT TO
		// in order to define for whom to send email message
		RecipientCheckers: []msmtpd.RecipientChecker{
			func(tr *msmtpd.Transaction, recipient *mail.Address) error {
				if strings.HasPrefix(recipient.Address, "info@") {
					return msmtpd.ErrorSMTP{
						Code:    535,
						Message: "Just stop it, please",
					}
				}
				return nil
			},
			recipient.AcceptMailForDomainsOrAddresses([]string{
				"example.org", // we accept all recipients from example.org domain
			}, []string{
				"anatolij@vodolaz095.ru", // we accept all emails for Anatolij (but it can have consequences)
			}),
		},
		// DataCheckers are called on message body to ensure it is properly formatted ham email
		// message according to RFC 5322 and RFC 6532.
		DataCheckers: []msmtpd.DataChecker{
			// at least message has minimal headers required
			data.CheckHeaders(data.DefaultHeadersToRequire),

			// this checker silently adds boss email where hidden
			// copies of messages are send using Transaction.Aliases
			func(tr *msmtpd.Transaction) error {
				for i := range tr.RcptTo {
					tr.Aliases = append(tr.Aliases, tr.RcptTo[i])
				}
				tr.Aliases = append(tr.Aliases, mail.Address{
					Name:    "Big Brother",
					Address: "big.brother@example.org",
				})
				return nil
			},
		},
		// DataHandlers are actual message delivery to persistent storage
		DataHandlers: []msmtpd.DataHandler{
			// we try to deliver via 3rd party proxy, note that
			// SMTP proxy code understands Transaction.Aliases
			deliver.ViaSMTPProxy(proxyOptions),
		},
		// CloseHandlers are called when client closes connection, they can be used
		// to, for example, record connection data in database or save metrics
		CloseHandlers: []msmtpd.CloseHandler{
			func(tr *msmtpd.Transaction) error {
				tr.LogInfo("Closing connection. Karma is %d", tr.Karma())
				return nil // error means nothing here, to be honest, connection is closed
			},
		},
		Logger: &logger,
	}

	err = server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}
