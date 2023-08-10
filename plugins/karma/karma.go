package karma

import (
	"fmt"

	"github.com/vodolaz095/msmtpd"
)

// DefaultInitialHate shows how much we respect 1st Law of Moses by hating strangers
const DefaultInitialHate = 10

// DefaultHateLimit defines how many bad things client can do until transaction will be marked as bad
const DefaultHateLimit = -4

// DefaultKarmaLimit is difference between good and bad connections to allow client to connect
const DefaultKarmaLimit = -5

// Good things giving karma points:
// 3 - HELO passed
// 3 - EHLO passed
// 3 - STARTTLS passed
// 3 - MAIL FROM passed
// 3 - RCP TO passed
// 3 - DataCheckers passed
// 18 = maximum good karma

// Bad things taking away karma points
// tlsHandshakeFailedHate = 1
// missingParameterPenalty = 1
// unknownCommandPenalty = 2
// tooManyRecipientsPenalty = 5
// malformedMessagePenalty = 5
// tooBigMessagePenalty = 5
// unknownRecipientPenalty = 1

// Handler is struct exposing Checkers for karma
type Handler struct {
	// InitialHate is negative karma given to new connections, if client issued commands properly, karma is improved
	InitialHate uint
	// HateLimit is how low karma can fall before we mark this transaction as bad
	HateLimit int

	// KarmaLimit is difference between number of good and bad connections. If KarmaLimit is -3, and client performed
	// 15 good connections and 17 bad connections, current karma will be 15-17=-2 and connection will be allowed
	KarmaLimit int

	// Storage defines interface for persistent (mainly) storage for Karma
	Storage Storage
}

// ConnectionChecker checks karma of remote IP address using data from Storage
func (kh *Handler) ConnectionChecker(tr *msmtpd.Transaction) (err error) {
	err = kh.Storage.Ping(tr.Context())
	if err != nil {
		tr.LogError(err, "while pinging karma storage")
		return msmtpd.ErrorSMTP{
			Code:    451,
			Message: "temporary errors, please, try again later",
		}
	}
	tr.Hate(int(kh.InitialHate))
	karma, err := kh.Storage.Get(tr)
	if err != nil {
		tr.LogError(err, fmt.Sprintf("while extracting transaction %s karma from storage", tr.ID))
		return msmtpd.ErrorSMTP{
			Code:    451,
			Message: "temporary errors, please, try again later",
		}
	}
	if karma > kh.KarmaLimit {
		tr.LogInfo("network address %s has acceptable karma %v for limit %v", tr.Addr, karma, kh.KarmaLimit)
		return nil
	}
	tr.LogWarn("network address %s has bad karma %v for limit %v", tr.Addr, karma, kh.KarmaLimit)
	return msmtpd.ErrorSMTP{
		Code:    521,
		Message: "FUCK OFF!", // lol
	}
}

// CloseHandler saves Transaction Karma into Storage after connection is finished
func (kh *Handler) CloseHandler(tr *msmtpd.Transaction) (err error) {
	var isGood bool
	if tr.Karma() > kh.HateLimit {
		tr.LogDebug("preparing to save transaction karma of %v as good", tr.Karma())
		isGood = true
		err = kh.Storage.SaveGood(tr)
	} else {
		tr.LogDebug("preparing to save transaction karma of %v as bad", tr.Karma())
		err = kh.Storage.SaveBad(tr)
	}
	if err != nil {
		tr.LogError(err, fmt.Sprintf("while saving transaction %s karma %v", tr.ID, tr.Karma()))
	} else {
		if isGood {
			tr.LogInfo("Transaction is saved as GOOD with love %v", tr.Karma())
		} else {
			tr.LogInfo("Transaction is saved as BAD with love %v", tr.Karma())
		}
	}
	return
}
