package data

import (
	"testing"
	"time"

	"msmtpd"
)

func TestBodyParseAndCheckHeadersOK(t *testing.T) {
	validMessage := `Date: Sun, 11 Jun 2023 19:49:29 +0300
To: scuba@vodolaz095.ru
From: scuba@vodolaz095.ru
Subject: test Sun, 11 Jun 2023 19:49:29 +0300
Message-Id: <20230611194929.017435@localhost>
X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/

This is a test mailing
`

	transaction := msmtpd.Transaction{
		ID:         "TestBodyParseAndCheckHeadersOK",
		StartedAt:  time.Now(),
		ServerName: "",
		Body:       []byte(validMessage),
	}
	handler := ParseBodyAndCheckHeaders(DefaultHeadersToRequire)

	err := handler(&transaction)
	if err != nil {
		t.Errorf("%s : while checking properly formated message", err)
	}
}

func TestBodyParseAndCheckHeadersMalformed(t *testing.T) {
	malformedMessage := `This is a test mailing`

	transaction := msmtpd.Transaction{
		ID:         "TestBodyParseAndCheckHeadersOK",
		StartedAt:  time.Now(),
		ServerName: "",
		Body:       []byte(malformedMessage),
	}
	handler := ParseBodyAndCheckHeaders(DefaultHeadersToRequire)

	err := handler(&transaction)
	if err != nil {
		typecasted, ok := err.(msmtpd.ErrorSMTP)
		if !ok {
			t.Errorf("error is not of kind msmtpd.ErrorSMTP")
		}
		if typecasted.Code != 521 {
			t.Errorf("wrong status code %v", typecasted.Code)
		}
		if typecasted.Message != complain {
			t.Errorf("wrong status message %s instead of %s", typecasted.Message, complain)
		}
		if typecasted.Error() != "521 "+complain {
			t.Errorf("wrong error message %s instead of 521 %s", typecasted.Error(), complain)
		}
	} else {
		t.Errorf("error not thrown while checking malformed message")
	}
}

func TestBodyParseAndCheckHeadersMissingHeaders(t *testing.T) {
	messageWithoutFrom := `Date: Sun, 11 Jun 2023 19:49:29 +0300
To: scuba@vodolaz095.ru
Subject: test Sun, 11 Jun 2023 19:49:29 +0300
Message-Id: <20230611194929.017435@localhost>
X-Mailer: swaks v20190914.0 jetmore.org/john/code/swaks/

This is a test mailing
`

	transaction := msmtpd.Transaction{
		ID:         "TestBodyParseAndCheckHeadersOK",
		StartedAt:  time.Now(),
		ServerName: "",
		Body:       []byte(messageWithoutFrom),
	}
	handler := ParseBodyAndCheckHeaders(DefaultHeadersToRequire)

	err := handler(&transaction)
	if err != nil {
		typecasted, ok := err.(msmtpd.ErrorSMTP)
		if !ok {
			t.Errorf("error is not of kind msmtpd.ErrorSMTP")
		}
		if typecasted.Code != 521 {
			t.Errorf("wrong status code %v", typecasted.Code)
		}
		if typecasted.Message != complain {
			t.Errorf("wrong status message %s instead of %s", typecasted.Message, complain)
		}
		if typecasted.Error() != "521 "+complain {
			t.Errorf("wrong error message %s instead of 521 %s", typecasted.Error(), complain)
		}
	} else {
		t.Errorf("error not thrown while checking malformed message")
	}
}
