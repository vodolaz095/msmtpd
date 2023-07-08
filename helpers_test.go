package msmptd

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
	cmd := parseLine("HELO hostname")
	if cmd.action != "HELO" {
		t.Fatalf("unexpected action: %s", cmd.action)
	}
	if len(cmd.fields) != 2 {
		t.Fatalf("unexpected fields length: %d", len(cmd.fields))
	}
	if len(cmd.params) != 1 {
		t.Fatalf("unexpected params length: %d", len(cmd.params))
	}
	if cmd.params[0] != "hostname" {
		t.Fatalf("unexpected value for param 0: %v", cmd.params[0])
	}
	cmd = parseLine("DATA")
	if cmd.action != "DATA" {
		t.Fatalf("unexpected action: %s", cmd.action)
	}
	if len(cmd.fields) != 1 {
		t.Fatalf("unexpected fields length: %d", len(cmd.fields))
	}
	if cmd.params != nil {
		t.Fatalf("unexpected params: %v", cmd.params)
	}
	cmd = parseLine("MAIL FROM:<test@example.org>")
	if cmd.action != "MAIL" {
		t.Fatalf("unexpected action: %s", cmd.action)
	}
	if len(cmd.fields) != 2 {
		t.Fatalf("unexpected fields length: %d", len(cmd.fields))
	}
	if len(cmd.params) != 2 {
		t.Fatalf("unexpected params length: %d", len(cmd.params))
	}
	if cmd.params[0] != "FROM" {
		t.Fatalf("unexpected value for param 0: %v", cmd.params[0])
	}
	if cmd.params[1] != "<test@example.org>" {
		t.Fatalf("unexpected value for param 1: %v", cmd.params[1])
	}
}

func TestParseLineMailformedMAILFROM(t *testing.T) {
	cmd := parseLine("MAIL FROM: <test@example.org>")
	if cmd.action != "MAIL" {
		t.Fatalf("unexpected action: %s", cmd.action)
	}
	if len(cmd.fields) != 2 {
		t.Fatalf("unexpected fields length: %d", len(cmd.fields))
	}
	if len(cmd.params) != 2 {
		t.Fatalf("unexpected params length: %d", len(cmd.params))
	}
	if cmd.params[0] != "FROM" {
		t.Fatalf("unexpected value for param 0: %v", cmd.params[0])
	}
	if cmd.params[1] != "<test@example.org>" {
		t.Fatalf("unexpected value for param 1: %v", cmd.params[1])
	}
}

func TestMask(t *testing.T) {
	masked := mask("thisIsNotAPassword")
	if masked != "t****" {
		t.Errorf("mask not works - %s", masked)
	}
}
