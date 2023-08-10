package deliver

import (
	"crypto/tls"
	"net"
	"net/smtp"

	"github.com/vodolaz095/msmtpd"
)

// SMTPProxyOptions used to configure ViaSMTPProxy
type SMTPProxyOptions struct {
	// Network defines how we dial SMTP Proxy, can be tcp or unix
	Network string
	// Address defines where remote SMTP Proxy is available
	Address string
	// HELO sets greeting for remote server
	HELO string
	// TLS, if not null, enables STARTTLS with this options
	TLS *tls.Config
	// Auth sets authentication mechanism for proxy backend
	Auth smtp.Auth
	// MailFrom defines senders override, if empty, value from Transaction.MailFrom will be used
	MailFrom string
	// RcptTo defines recipients' override, if empty, values from Transaction.Aliases or Transaction.RcptTo will be used
	RcptTo []string
}

var errProxyMalfunction = msmtpd.ErrorSMTP{
	Code:    451,
	Message: "temporary errors, please, try again later",
}

// ViaSMTPProxy adds DataHandler that performs delivery via 3rd party SMTP server
func ViaSMTPProxy(opts SMTPProxyOptions) msmtpd.DataHandler {
	return func(tr *msmtpd.Transaction) error {
		var i int
		var recipientsFound bool
		conn, err := net.Dial(opts.Network, opts.Address)
		if err != nil {
			tr.LogError(err, "error dialing SMTP backend")
			return errProxyMalfunction
		}
		client, err := smtp.NewClient(conn, opts.HELO)
		if err != nil {
			tr.LogError(err, "error making client to SMTP backend")
			return errProxyMalfunction
		}
		tr.LogDebug("Connection to SMTP backend %s is established via %s", opts.Address, opts.Network)
		err = client.Hello(opts.HELO)
		if err != nil {
			tr.LogError(err, "error making HELO/EHLO to SMTP backend")
			return errProxyMalfunction
		}
		tr.LogDebug("HELO/EHLO %s passed to SMTP backend", opts.HELO)
		if opts.TLS != nil {
			tr.LogDebug("Starting TLS to SMTP backend...")
			err = client.StartTLS(opts.TLS)
			if err != nil {
				tr.LogError(err, "error making STARTTLS to smtp backend")
				return errProxyMalfunction
			}
		}
		if opts.Auth != nil {
			tr.LogDebug("Starting Authorization to SMTP backend")
			err = client.Auth(opts.Auth)
			if err != nil {
				tr.LogError(err, "error performing authorization for smtp backend")
				return errProxyMalfunction
			}
			tr.LogDebug("Authorization to SMTP backend is passed")
		}

		if opts.MailFrom != "" {
			tr.LogDebug("Sending `MAIL FROM <%s>` like options says", opts.MailFrom)
			err = client.Mail(opts.MailFrom)
		} else {
			tr.LogDebug("Sending `MAIL FROM <%s>` from transaction", tr.MailFrom.Address)
			err = client.Mail(tr.MailFrom.Address)
		}
		if err != nil {
			tr.LogError(err, "error making MAILFROM to smtp backend")
			return errProxyMalfunction
		}
		if opts.RcptTo != nil {
			for i = range opts.RcptTo {
				tr.LogDebug("Sending `RCPT TO <%s>` from options...", opts.RcptTo[i])
				err = client.Rcpt(opts.RcptTo[i])
				if err != nil {
					tr.LogWarn("proxy recipient override %s is not accepted", opts.RcptTo[i])
				} else {
					tr.LogDebug("Sending `RCPT TO <%s>` accepted!", opts.RcptTo[i])
					recipientsFound = true
				}
			}
		} else {
			if tr.Aliases != nil {
				for i = range tr.Aliases {
					tr.LogDebug("Sending `RCPT TO <%s>` from aliases...", tr.Aliases[i].Address)
					err = client.Rcpt(tr.Aliases[i].Address)
					if err != nil {
						tr.LogWarn("original alias %s is not accepted", tr.Aliases[i].Address)
					} else {
						tr.LogDebug("Sending `RCPT TO <%s>` accepted!", tr.Aliases[i].Address)
						recipientsFound = true
					}
				}

			} else {
				if tr.RcptTo != nil {
					for i = range tr.RcptTo {
						tr.LogDebug("Sending `RCPT TO <%s>` from RCPT TO provided by client...", tr.RcptTo[i].Address)
						err = client.Rcpt(tr.RcptTo[i].Address)
						if err != nil {
							tr.LogWarn("original recipient %s is not accepted", tr.RcptTo[i].Address)
						} else {
							tr.LogDebug("Sending `RCPT TO <%s>` accepted!", tr.RcptTo[i].Address)
							recipientsFound = true
						}
					}
				}
			}
		}
		if !recipientsFound {
			tr.LogWarn("no recipients found")
			return UnknownRecipientError
		}
		wc, err := client.Data()
		if err != nil {
			tr.LogError(err, "error making DATA to smtp backend")
			return errProxyMalfunction
		}
		tr.LogDebug("DATA started...")
		i, err = wc.Write(tr.Body)
		if err != nil {
			tr.LogError(err, "error writing body to smtp backend")
			return errProxyMalfunction
		}
		tr.LogDebug("%v bytes of DATA is written, closing...", i)
		err = wc.Close()
		if err != nil {
			tr.LogError(err, "error closing data to smtp backend")
			return errProxyMalfunction
		}
		tr.LogDebug("DATA closed...")
		err = client.Close()
		if err != nil {
			tr.LogError(err, "error closing connection to smtp backend")
			return errProxyMalfunction
		}
		tr.LogInfo("Message of %v bytes is proxied to smtp backend %s", i, opts.Address)
		return nil
	}
}
