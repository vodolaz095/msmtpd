package msmtpd

import (
	"bytes"
	"encoding/base64"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

func (t *Transaction) handleAUTH(cmd command) {
	var mechanism, username, password string
	if len(cmd.fields) < 2 {
		t.reply(502, "Invalid syntax.")
		t.Hate(missingParameterPenalty)
		return
	}
	if t.server.Authenticator == nil {
		t.reply(502, "AUTH not supported.")
		t.Hate(missingParameterPenalty)
		return
	}
	if t.HeloName == "" {
		t.reply(502, "Please introduce yourself first.")
		t.Hate(missingParameterPenalty)
		return
	}

	if !t.Encrypted {
		t.reply(502, "Cannot AUTH in plain text mode. Use STARTTLS.")
		t.Hate(missingParameterPenalty)
		return
	}
	mechanism = strings.ToUpper(cmd.fields[1])
	switch mechanism {
	case "PLAIN":
		auth := ""
		if len(cmd.fields) < 3 {
			t.reply(334, "Give me your credentials")
			if !t.scanner.Scan() {
				return
			}
			auth = t.scanner.Text()
		} else {
			auth = cmd.fields[2]
		}
		data, err := base64.StdEncoding.DecodeString(auth)
		if err != nil {
			t.Hate(missingParameterPenalty)
			t.reply(502, "Couldn't decode your credentials")
			return
		}
		parts := bytes.Split(data, []byte{0})
		if len(parts) != 3 {
			t.Hate(missingParameterPenalty)
			t.reply(502, "Couldn't decode your credentials")
			return
		}
		username = string(parts[1])
		password = string(parts[2])
		break
	case "LOGIN":
		encodedUsername := ""
		if len(cmd.fields) < 3 {
			t.reply(334, "VXNlcm5hbWU6") // `Username:`
			if !t.scanner.Scan() {
				return
			}
			encodedUsername = t.scanner.Text()
		} else {
			encodedUsername = cmd.fields[2]
		}
		byteUsername, err := base64.StdEncoding.DecodeString(encodedUsername)
		if err != nil {
			t.Hate(missingParameterPenalty)
			t.reply(502, "Couldn't decode your credentials")
			return
		}
		t.reply(334, "UGFzc3dvcmQ6") // `Password:`
		if !t.scanner.Scan() {
			return
		}
		bytePassword, err := base64.StdEncoding.DecodeString(t.scanner.Text())
		if err != nil {
			t.reply(502, "Couldn't decode your credentials")
			return
		}
		username = string(byteUsername)
		password = string(bytePassword)
		break
	default:
		t.LogDebug("unknown authentication mechanism: %s", mechanism)
		t.reply(502, "Unknown authentication mechanism")
		return
	}
	t.LogDebug("Trying to authorise %s with password %s using mechanism %s",
		username, mask(password), mechanism,
	)
	t.Span.SetAttributes(attribute.String("username", username))
	t.Span.SetAttributes(attribute.String("password", mask(password)))
	err := t.server.Authenticator(t, username, password)
	if err != nil {
		t.error(err)
		t.Span.RecordError(err)
		return
	}
	t.Username = username
	t.Password = password
	t.reply(235, "OK, you are now authenticated")
	return
}
