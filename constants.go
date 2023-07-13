package msmtpd

import "crypto/tls"

const timeFormatForHeaders = "Mon, 02 Jan 2006 15:04:05 -0700 (MST)"

const lineLength = 76

// Karma related
const tlsHandshakeFailedHate = 1
const missingParameterPenalty = 1
const unknownCommandPenalty = 2

var TlsVersions = map[uint16]string{
	tls.VersionSSL30: "SSL3.0",
	tls.VersionTLS10: "TLS1.0",
	tls.VersionTLS11: "TLS1.1",
	tls.VersionTLS12: "TLS1.2",
	tls.VersionTLS13: "TLS1.3",
}
