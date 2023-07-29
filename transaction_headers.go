package msmtpd

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

/*
 * Header manipulation
 */

// AddHeader adds header, it should be called before AddReceivedLine, since it adds
// header to the top
func (t *Transaction) AddHeader(name, value string) {
	t.LogDebug("Adding header `%s: %s`", name, value)
	line := wrap([]byte(fmt.Sprintf("%s: %s\r\n", name, value)))
	t.Body = append(t.Body, line...)
	// Move the new newly added header line up front
	copy(t.Body[len(line):], t.Body[0:len(t.Body)-len(line)])
	copy(t.Body, line)
}

// AddReceivedLine prepends a Received header to the Data
func (t *Transaction) AddReceivedLine() {
	tlsDetails := ""
	if t.TLS != nil {
		version := "unknown"
		if val, ok := TLSVersions[t.TLS.Version]; ok {
			version = val
		}
		cipher := tls.CipherSuiteName(t.TLS.CipherSuite)
		tlsDetails = fmt.Sprintf(
			"\r\n\t(version=%s cipher=%s);",
			version,
			cipher,
		)
	}
	peerIP := ""
	if addr, ok := t.Addr.(*net.TCPAddr); ok {
		peerIP = addr.IP.String()
	}
	line := wrap([]byte(fmt.Sprintf(
		"Received: from %s ([%s]) by %s with %s;%s\r\n\t%s\r\n",
		t.HeloName,
		peerIP,
		t.ServerName,
		t.Protocol,
		tlsDetails,
		time.Now().Format(timeFormatForHeaders),
	)))
	t.Body = append(t.Body, line...)
	// Move the new Received line up front
	copy(t.Body[len(line):], t.Body[0:len(t.Body)-len(line)])
	copy(t.Body, line)
}
