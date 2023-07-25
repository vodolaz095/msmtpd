package internal

import "net/textproto"

func DoCommand(c *textproto.Conn, expectedCode int, format string, args ...interface{}) error {
	id, err := c.Cmd(format, args...)
	if err != nil {
		return err
	}
	c.StartResponse(id)
	_, _, err = c.ReadResponse(expectedCode)
	c.EndResponse(id)
	return err
}
