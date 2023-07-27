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
		t.reply(502, "malformed proxy command")
		return
	}
	var (
		newAddr    net.IP = nil
		newTCPPort uint64 = 0
		err        error
	)
	switch cmd.fields[1] {
	case "TCP4":
		break
	case "TCP6":
		break
	default:
		t.reply(502, "unable to decode proxy protocol - only TCP4/TCP6 is supported")
		return
	}

	newAddr = net.ParseIP(cmd.fields[2])
	if newAddr == nil {
		t.reply(502, "malformed network address")
		return
	}
	newTCPPort, err = strconv.ParseUint(cmd.fields[4], 10, 16)
	if err != nil {
		t.reply(502, "malformed port in proxy command")
		return
	}
	tcpAddr, ok := t.Addr.(*net.TCPAddr)
	if !ok {
		t.reply(502, "unsupported network connection")
		return
	}
	if newAddr != nil {
		tcpAddr.IP = newAddr
	}

	if newTCPPort != 0 {
		tcpAddr.Port = int(newTCPPort)
	}
	t.LogInfo("Proxy processed: new address - %s:%v",
		tcpAddr.IP, tcpAddr.Port,
	)
	t.Addr = tcpAddr
	t.welcome()
}
