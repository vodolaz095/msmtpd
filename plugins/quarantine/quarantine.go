package quarantine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/vodolaz095/msmtpd"
)

// FlagName is name of flag used to mark transaction's message as being quarantined
const FlagName = "quarantine"

// MoveToDirectory saves messages of flagged by FlagName transactions into directory using pattern directory/YYYY/MM/DD/{transactionID}.eml
func MoveToDirectory(directory string) msmtpd.DataHandler {
	err := os.MkdirAll(directory, 0755)
	if err != nil {
		log.Fatalf("%s : while making MoveToDirectory directory at %s", err, directory)
	}
	return func(tr *msmtpd.Transaction) error {
		if !tr.IsFlagSet(FlagName) {
			tr.LogDebug("MoveToDirectory flag is not set")
			return nil
		}
		dir := filepath.Join(directory,
			tr.StartedAt.Format("2006"),
			tr.StartedAt.Format("01"),
			tr.StartedAt.Format("02"),
		)
		trErr := os.MkdirAll(dir, 0755)
		if trErr != nil {
			tr.LogError(trErr, fmt.Sprintf("while creating quarantine directory at %s", dir))
			return msmtpd.ErrorSMTP{
				Code:    452,
				Message: "Requested action not taken: insufficient system storage",
			}
		}
		name := filepath.Join(dir, tr.ID+".eml")
		f, trErr := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if trErr != nil {
			tr.LogError(trErr, fmt.Sprintf("while creating quarantine file at %s", name))
			return msmtpd.ErrorSMTP{
				Code:    452,
				Message: "Requested action not taken: insufficient system storage",
			}
		}
		_, trErr = f.Write(tr.Body)
		if trErr != nil {
			tr.LogError(trErr, fmt.Sprintf("while writing quarantine file at %s", name))
			return msmtpd.ErrorSMTP{
				Code:    452,
				Message: "Requested action not taken: insufficient system storage",
			}
		}
		trErr = f.Close()
		if trErr != nil {
			tr.LogError(trErr, fmt.Sprintf("while closing quarantine file at %s", name))
			return msmtpd.ErrorSMTP{
				Code:    452,
				Message: "Requested action not taken: insufficient system storage",
			}
		}
		tr.LogInfo("Message quarantined into %s", name)
		return nil
	}

}
