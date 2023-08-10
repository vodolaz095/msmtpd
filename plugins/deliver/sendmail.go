package deliver

import (
	"bytes"
	"log"
	"os/exec"
	"strings"

	"github.com/vodolaz095/msmtpd"
)

// good read - https://man.archlinux.org/man/sendmail.8.en

// SendmailOptions defines how we start sendmail executable
type SendmailOptions struct {
	// PathToExecutable
	PathToExecutable string
	// UseMinusTFlag
	UseMinusTFlag bool
}

// ViaSendmail sends email message via sendmail command
func ViaSendmail(opts *SendmailOptions) msmtpd.DataHandler {
	if opts.PathToExecutable == "" {
		executablePath, lErr := exec.LookPath("sendmail")
		if lErr != nil {
			log.Fatalf("%s : while finding path to sendmail executable", lErr)
		}
		opts.PathToExecutable = executablePath
	}
	return func(tr *msmtpd.Transaction) error {
		var err error
		var args []string
		var recipients []string

		tr.LogDebug("Preparing to start sendmail executable at %s...", opts.PathToExecutable)
		if tr.MailFrom.Name != "" {
			args = append(args, "-F "+tr.MailFrom.Name)
		}
		args = append(args, "-f "+tr.MailFrom.Address)
		if opts.UseMinusTFlag {
			args = append(args, "-t")
		} else {
			if len(tr.Aliases) > 0 {
				for i := range tr.Aliases {
					recipients = append(recipients, tr.Aliases[i].Address)
				}
			} else {
				for i := range tr.RcptTo {
					recipients = append(recipients, tr.RcptTo[i].Address)
				}
			}
			args = append(args, strings.Join(recipients, ","))
		}
		cmd := exec.CommandContext(tr.Context(), opts.PathToExecutable, args...)
		tr.LogDebug("Preparing to execute %s...", cmd.String())
		cmd.Stdin = bytes.NewBuffer(tr.Body)
		output, err := cmd.CombinedOutput()
		if err != nil {
			tr.LogError(err, "while executing sendmail command")
			return TemporaryError
		}
		tr.LogDebug("Sendmail output is %s", string(output))
		if cmd.ProcessState.Success() {
			return nil
		}
		return TemporaryError
	}
}
