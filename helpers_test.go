package msmtpd

import (
	"testing"
)

func TestWrap(t *testing.T) {
	cases := map[string]string{
		"foobar":         "foobar",
		"foobar quux":    "foobar quux",
		"foobar\r\n":     "foobar\r\n",
		"foobar\r\nquux": "foobar\r\nquux",
		"foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux":      "foobar quux foobar quux foobar quux foobar quux foobar quux foobar quux foobar\r\n\tquux foobar quux",
		"foobar quux foobar quux foobar quux foobar quux foobar quux foobar\r\n\tquux foobar quux foobar quux": "foobar quux foobar quux foobar quux foobar quux foobar quux foobar\r\n\tquux foobar quux foobar quux",
	}
	for k, v := range cases {
		if string(wrap([]byte(k))) != v {
			t.Fatal("Didn't wrap correctly.")
		}
	}
}

func TestParseLine(t *testing.T) {
	commandForParseLine := parseLine("HELO hostname")
	if commandForParseLine.action != "HELO" {
		t.Fatalf("unexpected action: %s", commandForParseLine.action)
	}
	if len(commandForParseLine.fields) != 2 {
		t.Fatalf("unexpected fields length: %d", len(commandForParseLine.fields))
	}
	if len(commandForParseLine.params) != 1 {
		t.Fatalf("unexpected params length: %d", len(commandForParseLine.params))
	}
	if commandForParseLine.params[0] != "hostname" {
		t.Fatalf("unexpected value for param 0: %v", commandForParseLine.params[0])
	}
	commandForParseLine = parseLine("DATA")
	if commandForParseLine.action != "DATA" {
		t.Fatalf("unexpected action: %s", commandForParseLine.action)
	}
	if len(commandForParseLine.fields) != 1 {
		t.Fatalf("unexpected fields length: %d", len(commandForParseLine.fields))
	}
	if commandForParseLine.params != nil {
		t.Fatalf("unexpected params: %v", commandForParseLine.params)
	}
	commandForParseLine = parseLine("MAIL FROM:<test@example.org>")
	if commandForParseLine.action != "MAIL" {
		t.Fatalf("unexpected action: %s", commandForParseLine.action)
	}
	if len(commandForParseLine.fields) != 2 {
		t.Fatalf("unexpected fields length: %d", len(commandForParseLine.fields))
	}
	if len(commandForParseLine.params) != 2 {
		t.Fatalf("unexpected params length: %d", len(commandForParseLine.params))
	}
	if commandForParseLine.params[0] != "FROM" {
		t.Fatalf("unexpected value for param 0: %v", commandForParseLine.params[0])
	}
	if commandForParseLine.params[1] != "<test@example.org>" {
		t.Fatalf("unexpected value for param 1: %v", commandForParseLine.params[1])
	}
}

func TestParseLineMalformedMAILFROM(t *testing.T) {
	commandForParseMalformedMailFrom := parseLine("MAIL FROM: <test@example.org>")
	if commandForParseMalformedMailFrom.action != "MAIL" {
		t.Fatalf("unexpected action: %s", commandForParseMalformedMailFrom.action)
	}
	if len(commandForParseMalformedMailFrom.fields) != 2 {
		t.Fatalf("unexpected fields length: %d", len(commandForParseMalformedMailFrom.fields))
	}
	if len(commandForParseMalformedMailFrom.params) != 2 {
		t.Fatalf("unexpected params length: %d", len(commandForParseMalformedMailFrom.params))
	}
	if commandForParseMalformedMailFrom.params[0] != "FROM" {
		t.Fatalf("unexpected value for param 0: %v", commandForParseMalformedMailFrom.params[0])
	}
	if commandForParseMalformedMailFrom.params[1] != "<test@example.org>" {
		t.Fatalf("unexpected value for param 1: %v", commandForParseMalformedMailFrom.params[1])
	}
}

func TestMask(t *testing.T) {
	masked := mask("thisIsNotAPassword")
	if masked != "t****" {
		t.Errorf("mask not works - %s", masked)
	}
}

func TestDecodeBase64EncodedSubject(t *testing.T) {
	var decoded string
	var err error
	cases := map[string]string{
		"=?UTF-8?B?0YHQvtC9INCh0LLQtdGC0LvQsNC90Ys=?=":                                                              "сон Светланы",
		"=?utf-8?b?RXh0ZXJuYWwgYW5zaWJsZSByZXBvcnQgLSDQstGB0ZEg0YDQsNCx0L7RgtCw0LXRgg==?=":                          "External ansible report - всё работает",
		"=?UTF-8?b?RXh0ZXJuYWwgYW5zaWJsZSByZXBvcnQgLSDQstGB0ZEg0YDQsNCx0L7RgtCw0LXRgg==?=":                          "External ansible report - всё работает",
		"=?utf-8?B?0JfQsNC00LDQudGC0LUg0LLQvtC/0YDQvtGBINC+INC/0YDQvtC40YE=?= =?utf-8?B?0YXQvtC20LTQtdC90LjQuCE=?=": "Задайте вопрос о происхождении!",
	}
	for k, v := range cases {
		decoded, err = decodeBase64EncodedSubject(k)
		if err != nil {
			t.Errorf("%s : while decoding `%s` into `%s`", err, k, v)
		} else {
			if decoded != v {
				t.Errorf("wrong data decoded - `%s` into `%s`", decoded, v)
			}
		}
	}
	erroneous, err := decodeBase64EncodedSubject("=?utf-8?B?этоНеbase64?=")
	if err != nil {
		if err.Error() != "illegal base64 data at input byte 0" {
			t.Errorf("%s : while decoding", err)
		}
	}
	if erroneous != "=?utf-8?B?этоНеbase64?=" {
		t.Error("should not modify")
	}
}
