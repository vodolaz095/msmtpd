package karma

import (
	"fmt"

	"msmtpd"
)

type Handler struct {
	HateLimit int
	Storage   Storage
}

func (kh *Handler) ConnectionChecker(tr *msmtpd.Transaction) (err error) {
	err = kh.Storage.Ping(tr.Context())
	if err != nil {
		tr.LogWarn("%s : while pinging karma storage", err)
	}
	karma, err := kh.Storage.Get(tr)
	if err != nil {
		tr.LogError(err, fmt.Sprintf("while extracting transaction %s karma from storage", tr.ID))
		return
	}
	if karma > kh.HateLimit {
		tr.LogDebug("network address %s has acceptable karma %v", tr.Addr, karma)
		return nil
	}

	tr.LogDebug("network address %s has bad karma %v", tr.Addr, karma)
	return msmtpd.ErrorSMTP{
		Code:    521,
		Message: "FUCK OFF!", // lol
	}
}

func (kh *Handler) CloseChecker(tr *msmtpd.Transaction) (err error) {
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
		tr.LogInfo("Transaction karma is saved")
	}
	return
}