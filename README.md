msmtpd
=======================

[![PkgGoDev](https://pkg.go.dev/badge/github.com/vodolaz095/msmtpd)](https://pkg.go.dev/github.com/vodolaz095/msmtpd?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/vodolaz095/msmtpd)](https://goreportcard.com/report/github.com/vodolaz095/msmtpd)
[![Go](https://github.com/vodolaz095/msmtpd/actions/workflows/go.yml/badge.svg)](https://github.com/vodolaz095/msmtpd/actions/workflows/go.yml)

Golang framework for building Simple Mail Transfer Protocol Daemons.

Фреймворк для создания почтовых серверов, написанный на Go.

Mirror - https://gitflic.ru/project/vodolaz095/msmtpd

Main features
================================

1. [Haraka hooks](https://haraka.github.io/core/Plugins#available-hooks) inspired CheckerFunc's being called 
   on different actions of client (connection, HELO/EHLO command, StartTLS)
2. `StartTLS`, `XClient`, `Proxy` command support
3. Easy to implement logger interface
4. Build-in [OpenTelemetry](https://opentelemetry.io/) support - see [dovecot_inbound](example%2Fdovecot_inbound)
   and [dovecot_outbound](example%2Fdovecot_outbound) examples
5. Lot of [plugins](plugins), including:
6. [Dovecot](plugins%2Fdovecot) plugin for authentication and LMTP mail delivery
7. [Rspamd](plugins%2Frspamd) plugin for blocking spam
8. Plugins to deliver via 3rd party [SMTP proxy](plugins%2Fdeliver%2Fsmtp_proxy.go), 
   [LMTP](plugins%2Fdeliver%2Flmtp.go) and [SendMail](plugins%2Fdeliver%2Fsendmail.go).
9. Experimental [Karma](plugins%2Fkarma) plugin to implement connection scoring (IP addresses making failed SMTP transactions will be blacklisted)
10. HELO/EHLO checkers, including complicated [ones](plugins%2Fhelo)
11. [Sender resolvable](plugins%2Fsender%2Fsender_resolvable.go) checker plugin to ensure sender's domain can accept our replies 

Examples / Примеры
================================

- [custom_logger](example%2Fcustom_logger)
- [dovecot_inbound](example%2Fdovecot_inbound)
- [dovecot_outbound](example%2Fdovecot_outbound)
- [metrics](example%2Fmetrics)
- [minimal](example%2Fminimal)
- [simple](example%2Fsimple)
- [smtp_proxy](example%2Fsmtp_proxy)

Inspiration / Источники вдохновения
=================================

- https://github.com/chrj/smtpd (часть кода)
- https://haraka.github.io/ (реализация функционала плагинов, общая идеология)
- https://github.com/albertito/chasquid/ (реализация проверки получателей и авторизации через сокеты Dovecot)

Minimal example / Минимальный пример
==================================

```go

package main

import (
	"log"

	"github.com/vodolaz095/msmtpd"
)

func main() {
	server := msmtpd.Server{
		Hostname:       "localhost",
		WelcomeMessage: "Do you believe in our God?",
	}

	err := server.ListenAndServe(":1025")
	if err != nil {
		log.Fatalf("%s : while starting server on 0.0.0.0:1025", err)
	}
}


```



Server log / Протокол работы сервера
=====================
```
2023/08/02 16:30:11 INFO [1cca484c18ebe494240b196bc60ee39c]: Accepting connection from [::1]:34142...
2023/08/02 16:30:11 INFO [1cca484c18ebe494240b196bc60ee39c]: EHLO <localhost> is accepted!
2023/08/02 16:30:11 INFO [1cca484c18ebe494240b196bc60ee39c]: MAIL FROM <sender@example.org> is checked by 0
SenderCheckers and accepted!
2023/08/02 16:30:11 INFO [1cca484c18ebe494240b196bc60ee39c]: Recipient <recipient@example.org> will be 1st one in
transaction!
2023/08/02 16:30:11 INFO [1cca484c18ebe494240b196bc60ee39c]: Subject: test Wed, 02 Aug 2023 16:30:11 +0300
2023/08/02 16:30:11 INFO [1cca484c18ebe494240b196bc60ee39c]: Body (264 bytes) checked by 0 DataCheckers successfully!
2023/08/02 16:30:11 WARN [1cca484c18ebe494240b196bc60ee39c]: Message silently discarded - no DataHandlers set...
2023/08/02 16:30:12 INFO [1cca484c18ebe494240b196bc60ee39c]: Closing transaction.
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Accepting connection from [::1]:50596...
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: EHLO <localhost> is accepted!
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: MAIL FROM <sender@example.org> is checked by 0
SenderCheckers and accepted!
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Recipient <recipient1@example.org> will be 1st one in
transaction!
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Recipient <recipient2@example.org> will be 2nd one in
transaction!
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Recipient <recipient3@example.org> will be 2nd one in
transaction!
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Recipient <recipient4@example.org> will be 4th one in
transaction!
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Subject: test Wed, 02 Aug 2023 16:30:31 +0300
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Body (334 bytes) checked by 0 DataCheckers successfully!
2023/08/02 16:30:31 WARN [4dfe768e04186309306b1af3194b05c4]: Message silently discarded - no DataHandlers set...
2023/08/02 16:30:31 INFO [4dfe768e04186309306b1af3194b05c4]: Closing transaction.
```

Client log while testing smtp server / Вывод клиента при тестировании почтового сервера
=====================

```shell
$ swaks --to recipient1@example.org,recipient2@example.org,recipient3@example.org,recipient4@example.org \
    --from sender@example.org \
    --server localhost --port 1025 \
    --timeout 600
```

```
=== Trying localhost:1025...
=== Connected to localhost.
<- 220 Do you believe in our God?
-> EHLO localhost
<- 250-localhost
<- 250-SIZE 10240000
<- 250-8BITMIME
<- 250 PIPELINING
-> MAIL FROM:<sender@example.org>
<- 250 Ok, it makes sense, go ahead please!
-> RCPT TO:<recipient1@example.org>
<- 250 It seems i can handle delivery for this recipient, i'll do my best!
-> RCPT TO:<recipient2@example.org>
<- 250 It seems i can handle delivery for this recipient, i'll do my best!
-> RCPT TO:<recipient3@example.org>
<- 250 It seems i can handle delivery for this recipient, i'll do my best!
-> RCPT TO:<recipient4@example.org>
<- 250 It seems i can handle delivery for this recipient, i'll do my best!
-> DATA
<- 354 Ok, you managed to talk me into accepting your message. Go on, end your data with <CR><LF>.<CR><LF>
-> Date: Wed, 02 Aug 2023 16:30:31 +0300
-> To: recipient1@example.org,recipient2@example.org,recipient3@example.org,recipient4@example.org
-> From: sender@example.org
-> Subject: test Wed, 02 Aug 2023 16:30:31 +0300
-> Message-Id: <20230802163031.059039@localhost>
-> X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/
->
-> This is a test mailing
->
->
-> .
<- 250 Thank you.
-> QUIT
<- 221 Farewell, my friend! Transaction 4dfe768e04186309306b1af3194b05c4 is finished
```



Support development / Поддержать разработку
====================
https://www.tinkoff.ru/rm/ostroumov.anatoliy2/4HFzm76801/
