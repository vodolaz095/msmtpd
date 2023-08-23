package msmtpd

import (
	"context"
	"log"
	"net"
	"net/mail"
	"strings"
	"time"
)

func Example() {
	// set main server context
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// configure logger
	logger := DefaultLogger{
		Logger: log.Default(),
		Level:  TraceLevel,
	}
	// configure server
	server := Server{
		Hostname:       "mx.example.org",
		WelcomeMessage: "mx.example.org ESMTP ready.",

		// limits
		ReadTimeout:    time.Minute,
		WriteTimeout:   time.Minute,
		DataTimeout:    3 * time.Minute,
		MaxConnections: 10,
		MaxMessageSize: 32 * 1024 * 1014, // 32 Mbytes
		MaxRecipients:  10,
		// you can configure DNS resolver being used for application
		Resolver: net.DefaultResolver,
		// can make things faster, but various HELO/EHLO checks will not work
		SkipResolvingPTR: false,
		// enable xclient extension, use with care! https://www.postfix.org/XCLIENT_README.html
		EnableXCLIENT: false,
		// enable proxy support, like Haraka allows for HaProxy https://haraka.github.io/core/HAProxy
		EnableProxyProtocol: false,

		// if you plan to make server listen TLS or use StartTLS command, here
		// you can configure it and require all sensitive data to be transferred via
		// encrypted channel
		TLSConfig: nil,
		ForceTLS:  false,

		// plug your custom logger here
		Logger: &logger,

		// main context of server and its cancel function
		Context: ctx,
		Cancel:  cancel,

		// ConnectionCheckers are called when client performed TCP connection
		ConnectionCheckers: []ConnectionChecker{
			func(tr *Transaction) error {
				tr.LogInfo("Client connects from %s...", tr.Addr.String())
				return nil
			},
		},
		// HeloCheckers are called when client tries to perform HELO/EHLO
		HeloCheckers: []HelloChecker{
			func(tr *Transaction) error {
				tr.LogInfo("Client HELO is %s", tr.HeloName)
				return nil
			},
			func(tr *Transaction) error {
				if tr.HeloName != "localhost" {
					tr.Hate(1) // i do not like being irritated
				} else {
					tr.SetFlag("localhost")
				}
				return nil
			},
		},
		// SenderCheckers are called when client provides MAIL FROM to define who sends email message
		SenderCheckers: []SenderChecker{
			func(tr *Transaction) error {
				tr.LogInfo("Somebody called %s tries to send message...", tr.MailFrom)
				return nil
			},
		},
		// RecipientCheckers are called each time client provides RCPT TO
		// in order to define for whom to send email message
		RecipientCheckers: []RecipientChecker{
			func(tr *Transaction, recipient *mail.Address) error {
				if strings.HasPrefix(recipient.Address, "info@") {
					return ErrorSMTP{
						Code:    535,
						Message: "Just stop it, please",
					}
				}
				return nil
			},
		},
		// DataCheckers are called on message body to ensure it is properly formatted ham email
		// message according to RFC 5322 and RFC 6532.
		DataCheckers: []DataChecker{
			func(tr *Transaction) error {
				if tr.Parsed.Header.Get("X-Priority") == "" {
					return ErrorSMTP{
						Code:    535,
						Message: "Please, provide priority for your message!",
					}
				}
				// Add header to message
				tr.AddHeader("Something", "interesting")
				return nil
			},
		},
		// DataHandlers do actual message delivery to persistent storage
		DataHandlers: []DataHandler{
			func(tr *Transaction) error {
				tr.LogInfo("We pretend we deliver %v bytes of message somehow", len(tr.Body))
				// set float64 fact about transaction
				tr.Incr("size", float64(len(tr.Body)))
				return ErrServiceNotAvailable
			},
		},
		// CloseHandlers are called when client closes connection, they can be used
		// to, for example, record connection data in database or save metrics
		CloseHandlers: []CloseHandler{
			func(tr *Transaction) error {
				tr.LogInfo("Closing connection. Karma is %d", tr.Karma())
				// reading string fact
				subject, found := tr.GetFact(SubjectFact)
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
	// start http endpoint for metrics scrapping.
	// Format explained here: https://prometheus.io/docs/instrumenting/exposition_formats/
	go func() {
		pmErr := server.StartPrometheusScrapperEndpoint(":3000", "/metrics")
		if pmErr != nil {
			log.Fatalf("%s : while starting metrics scrapper endpoint", pmErr)
		}
	}()
	// Actually, start server on 25th port
	err := server.ListenAndServe(":25")
	if err != nil {
		log.Fatalf("%s : while starting server on %s", err, server.Address())
	}
}
