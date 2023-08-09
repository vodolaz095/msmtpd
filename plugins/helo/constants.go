package helo

import "github.com/vodolaz095/msmtpd"

var complain = msmtpd.ErrorSMTP{
	Code:    521,
	Message: "I don't like the way you introduce yourself. Goodbye!",
}

// IsLocalAddressFlagName is flag name to mark local remote addresses
const IsLocalAddressFlagName = "addr_is_local"

// DefaultHateForReverseDNSMismatch is how much we punish by default for Reverse DNS mismatch
const DefaultHateForReverseDNSMismatch = 3
