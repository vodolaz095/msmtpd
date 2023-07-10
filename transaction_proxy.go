package msmtpd

import (
	"net"
	"strconv"
)

// Additional documentation:
// http://www.haproxy.org/download/1.8/doc/proxy-protocol.txt
// Example:  `PROXY TCP4 8.8.8.8 127.0.0.1 443 25`

func (t *Transaction) handlePROXY(cmd command) {
	t.LogTrace("Proxy command: %s", cmd.line)
	if !t.server.EnableProxyProtocol {
		t.reply(550, "Proxy Protocol not enabled")
		return
	}
	if len(cmd.fields) < 6 {
		t.reply(502, "Couldn't decode the command.")
		return
	}
	var (
		newAddr    net.IP = nil
		newTCPPort uint64 = 0
		err        error
	)
	newAddr = net.ParseIP(cmd.fields[2])

	newTCPPort, err = strconv.ParseUint(cmd.fields[4], 10, 16)
	if err != nil {
		t.reply(502, "Couldn't decode the command.")
		return
	}
	tcpAddr, ok := t.Addr.(*net.TCPAddr)
	if !ok {
		t.reply(502, "Unsupported network connection")
		return
	}
	if newAddr != nil {
		tcpAddr.IP = newAddr
	}

	if newTCPPort != 0 {
		tcpAddr.Port = int(newTCPPort)
	}
	t.LogDebug("Proxy processed: new address - %s:%v",
		tcpAddr.IP, tcpAddr.Port,
	)
	t.welcome()
}
