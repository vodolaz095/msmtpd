package karma

import (
	"fmt"

	"msmtpd"
)

// DefaultLoveRequired are good karma points required for transaction to be considered ok
const DefaultLoveRequired = 10

// 3 - HELO/EHLO
// 3 - MAIL FROM
// 3 - RCP TO
// 3 - DATA

// Handler is struct exposing Checkers for karma
type Handler struct {
	HateLimit int
	Storage   Storage
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
	karma, err := kh.Storage.Get(tr)
	if err != nil {
		tr.LogError(err, fmt.Sprintf("while extracting transaction %s karma from storage", tr.ID))
		return msmtpd.ErrorSMTP{
			Code:    451,
			Message: "temporary errors, please, try again later",
		}
	}
	if karma > kh.HateLimit {
		tr.LogInfo("network address %s has acceptable karma %v for limit %v", tr.Addr, karma, kh.HateLimit)
		return nil
	}
	tr.LogWarn("network address %s has bad karma %v for limit %v", tr.Addr, karma, kh.HateLimit)
	return msmtpd.ErrorSMTP{
		Code:    521,
		Message: "FUCK OFF!", // lol
	}
}

// CloseHandler saves Transaction Karma into Storage after connection is finished
func (kh *Handler) CloseHandler(tr *msmtpd.Transaction) (err error) {
	if tr.Karma() > kh.HateLimit {
		tr.LogDebug("preparing to save transaction karma of %v as good", tr.Karma())
		err = kh.Storage.SaveGood(tr)
	} else {
		tr.LogDebug("preparing to save transaction karma of %v as bad", tr.Karma())
		err = kh.Storage.SaveBad(tr)
	}
	if err != nil {
		tr.LogError(err, fmt.Sprintf("while saving transaction %s karma %v", tr.ID, tr.Karma()))
	} else {
		tr.LogInfo("Transaction karma %v is saved", tr.Karma())
	}
	return
}
