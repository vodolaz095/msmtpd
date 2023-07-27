package msmtpd

import (
	"net"
	"strconv"
	"strings"
)

func (t *Transaction) handleXCLIENT(cmd command) {
	if len(cmd.fields) < 2 {
		t.reply(502, "Invalid syntax.")
		return
	}
	if !t.server.EnableXCLIENT {
		t.reply(550, "XCLIENT not enabled")
		return
	}
	var (
		newHeloName, newUsername string
		newProto                 Protocol
		newAddr                  net.IP
		newTCPPort               uint64
	)
	for _, item := range cmd.fields[1:] {
		parts := strings.Split(item, "=")
		if len(parts) != 2 {
			t.reply(502, "Couldn't decode the command.")
			return
		}
		name := parts[0]
		value := parts[1]
		switch name {
		case "NAME":
			// Unused in smtpd package
			continue
		case "HELO":
			newHeloName = value
			continue
		case "ADDR":
			newAddr = net.ParseIP(value)
			continue
		case "PORT":
			var err error
			newTCPPort, err = strconv.ParseUint(value, 10, 16)
			if err != nil {
				t.reply(502, "Couldn't decode the command.")
				return
			}
			continue
		case "LOGIN":
			newUsername = value
			continue
		case "PROTO":
			if value == "SMTP" {
				newProto = SMTP
			} else if value == "ESMTP" {
				newProto = ESMTP
			}
			continue
		default:
			t.reply(502, "Couldn't decode the command.")
			return
		}
	}
	tcpAddr, ok := t.Addr.(*net.TCPAddr)
	if !ok {
		t.reply(502, "Unsupported network connection")
		return
	}
	if newHeloName != "" {
		t.HeloName = newHeloName
	}
	if newAddr != nil {
		tcpAddr.IP = newAddr
	}
	if newTCPPort != 0 {
		tcpAddr.Port = int(newTCPPort)
	}
	if newUsername != "" {
		t.Username = newUsername
	}
	if newProto != "" {
		t.Protocol = newProto
	}
	if newAddr != nil && newTCPPort != 0 {
		t.Addr = tcpAddr
	}
	t.welcome()
}
