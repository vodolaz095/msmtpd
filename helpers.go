package msmtpd

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// wrap a byte slice paragraph for use in SMTP header
func wrap(sl []byte) []byte {
	length := 0
	for i := 0; i < len(sl); i++ {
		if length > lineLength && sl[i] == ' ' {
			sl = append(sl, 0, 0)
			copy(sl[i+2:], sl[i:])
			sl[i] = '\r'
			sl[i+1] = '\n'
			sl[i+2] = '\t'
			i += 2
			length = 0
		}
		if sl[i] == '\n' {
			length = 0
		}
		length++
	}
	return sl
}

// parseLine used to parse string into command
func parseLine(line string) (cmd command) {
	cmd.line = line
	cmd.fields = strings.Fields(line)
	if len(cmd.fields) > 0 {
		cmd.action = strings.ToUpper(cmd.fields[0])
		if len(cmd.fields) > 1 {
			// Account for some clients breaking the standard and having
			// an extra whitespace after the ':' character. Example:
			// MAIL FROM: <test@example.org>
			//
			// Should be:
			//
			// MAIL FROM:<test@example.org>
			//
			// Thus, we add a check if the second field ends with ':'
			// and appends the rest of the third field.

			if cmd.fields[1][len(cmd.fields[1])-1] == ':' && len(cmd.fields) > 2 {
				cmd.fields[1] = cmd.fields[1] + cmd.fields[2]
				cmd.fields = cmd.fields[0:2]
			}
			cmd.params = strings.Split(cmd.fields[1], ":")
		}
	}
	return
}

func mask(input string) string {
	return string(input[0]) + "****"
}

// getRandomID gets random hex encoded id
func getRandomID() (id string, err error) {
	b := make([]byte, 10)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	id = hex.EncodeToString(b)
	return
}
