package msmtpd

type command struct {
	line   string
	action string
	fields []string
	params []string
}
