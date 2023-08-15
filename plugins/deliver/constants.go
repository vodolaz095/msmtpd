package deliver

import "time"

// DefaultTimeout is default timeout
const DefaultTimeout = 5 * time.Second

// DiscardFlag makes, while being set, message to be silently discarded by all compatible msmtpd.DataHandler's
const DiscardFlag = "discard"
