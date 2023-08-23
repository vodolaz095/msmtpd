package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vodolaz095/msmtpd"
)

// This example shows how to use custom logger

type customLogger struct {
	*log.Logger
}

func (cl *customLogger) logf(transaction *msmtpd.Transaction, format string, args ...any) string {
	return fmt.Sprintf("[%s %s]", transaction.ID, transaction.Addr.String()) +
		" " + fmt.Sprintf(format, args...)
}
func (cl *customLogger) Tracef(transaction *msmtpd.Transaction, format string, args ...any) {
	cl.Println("TRACE ", cl.logf(transaction, format, args...))
}
func (cl *customLogger) Debugf(transaction *msmtpd.Transaction, format string, args ...any) {
	cl.Println("DEBUG ", cl.logf(transaction, format, args...))
}
func (cl *customLogger) Infof(transaction *msmtpd.Transaction, format string, args ...any) {
	cl.Println("INFO  ", cl.logf(transaction, format, args...))
}
func (cl *customLogger) Warnf(transaction *msmtpd.Transaction, format string, args ...any) {
	cl.Println("WARN  ", cl.logf(transaction, format, args...))
}
func (cl *customLogger) Errorf(transaction *msmtpd.Transaction, format string, args ...any) {
	cl.Println("ERROR ", cl.logf(transaction, format, args...))
}
func (cl *customLogger) Fatalf(transaction *msmtpd.Transaction, format string, args ...any) {
	cl.Fatal("FATAL ", cl.logf(transaction, format, args...))
}

func main() {
	server := msmtpd.Server{
		Hostname:       "localhost",
		WelcomeMessage: "Do you believe in our God?",
		Logger: &customLogger{
			Logger: log.New(os.Stdout, "msmtpd ", log.Ldate|log.Ltime|log.Lmsgprefix),
		},
	}

	err := server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}

/*

Server output is something like

[vodolaz095@steel custom_logger]$ go run main.go
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Starting transaction f3229d44a4c813ba581dd9019638ef79 for [::1]:51992.
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] PTR addresses resolved for [::1]:51992 : [localhost]
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Accepting connection from [::1]:51992...
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Sending: 220 Do you believe in our God?
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Received: EHLO localhost
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Command received: EHLO localhost
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] EHLO <localhost> is received...
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] EHLO <localhost> is accepted!
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Sending: 250 PIPELINING
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Granting 3 love for transaction
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Setting counter karma to 3
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Received: MAIL FROM:<sender@example.org>
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Command received: MAIL FROM:<sender@example.org>
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Checking MAIL FROM <sender@example.org> by 0 SenderCheckers...
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] MAIL FROM <sender@example.org> is checked by 0 SenderCheckers and accepted!
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Sending: 250 Ok, it makes sense, go ahead please!
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Granting 3 love for transaction
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Incrementing karma by 3 from 3 to 6
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Received: RCPT TO:<recipient@example.org>
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Command received: RCPT TO:<recipient@example.org>
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Checking recipient <recipient@example.org> by 0 RecipientCheckers...
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Recipient <recipient@example.org> will be 1st one in transaction!
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Sending: 250 It seems i can handle delivery for this recipient, i'll do my best!
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Granting 3 love for transaction
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Incrementing karma by 3 from 6 to 9
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Received: DATA
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Command received: DATA
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] DATA is called...
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Sending: 354 Ok, you managed to talk me into accepting your message. Go on, end your data with <CR><LF>.<CR><LF>
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Adding header `MSMTPD-Transaction-Id: f3229d44a4c813ba581dd9019638ef79`
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Parsing message body with size 264...
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Message created on Wed, 23 Aug 2023 20:37:29 +0300 (MSK) - 734.159937ms ago
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Subject: test Wed, 23 Aug 2023 20:37:29 +0300
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Message body of 264 bytes is parsed, calling 0 DataCheckers on it
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Body (264 bytes) checked by 0 DataCheckers successfully!
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Granting 3 love for transaction
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Incrementing karma by 3 from 9 to 12
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Starting delivery by 0 DataHandlers...
2023/08/23 20:37:29 msmtpd WARN   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Message silently discarded - no DataHandlers set...
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Sending: 250 Thank you.
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Granting 3 love for transaction
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Incrementing karma by 3 from 12 to 15
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Received: QUIT
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Command received: QUIT
2023/08/23 20:37:29 msmtpd TRACE  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Sending: 221 Farewell, my friend! Transaction f3229d44a4c813ba581dd9019638ef79 is finished
2023/08/23 20:37:29 msmtpd DEBUG  [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Starting 0 close handlers...
2023/08/23 20:37:29 msmtpd INFO   [f3229d44a4c813ba581dd9019638ef79 [::1]:51992] Closing transaction f3229d44a4c813ba581dd9019638ef79.



*/
