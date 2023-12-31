package internal

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// MakeTestMessage makes test email message body suitable for unit tests
func MakeTestMessage(from string, to ...string) string {
	now := time.Now()
	buh := bytes.NewBufferString("Date: " + now.Format(time.RFC1123Z) + "\r\n")
	buh.WriteString("From: " + from + "\r\n")
	buh.WriteString("To: " + strings.Join(to, ",") + "\r\n")
	buh.WriteString(fmt.Sprintf("Subject: Test email send on %s\r\n", now.Format(time.RFC1123Z)))
	buh.WriteString(fmt.Sprintf("Message-Id: <%s@localhost>\r\n", now.Format("20060102150405")))
	buh.WriteString("\r\n")
	buh.WriteString(fmt.Sprintf("This is test message send from %s to %s on %s\r\n",
		from, strings.Join(to, ","), now.Format(time.Stamp),
	))
	return buh.String()
}
