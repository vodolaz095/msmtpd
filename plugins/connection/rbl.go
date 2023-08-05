package connection

// Good read
// https://multirbl.valli.org/

// SpamhauseReverseIPBlackLists lists to check IP against databases
var SpamhauseReverseIPBlackLists = []string{
	"pbl.spamhaus.org",
	"sbl.spamhaus.org",
	"sbl-xbl.spamhaus.org",
	"xbl.spamhaus.org",
	"zen.spamhaus.org",
}

// SpamEatingMonkeyReverseIPBlackLists lists to check IP against databases
var SpamEatingMonkeyReverseIPBlackLists = []string{
	"bl.spameatingmonkey.net",
	"backscatter.spameatingmonkey.net",
}

// SorbsReverseIPBlacklists lists to check IP against databases
var SorbsReverseIPBlacklists = []string{
	"dnsbl.sorbs.net",
	"problems.dnsbl.sorbs.net",
	"proxies.dnsbl.sorbs.net",
	"relays.dnsbl.sorbs.net",
	"safe.dnsbl.sorbs.net",
	"nomail.rhsbl.sorbs.net",
	"badconf.rhsbl.sorbs.net",
	"dul.dnsbl.sorbs.net",
	"zombie.dnsbl.sorbs.net",
	"block.dnsbl.sorbs.net",
	"escalations.dnsbl.sorbs.net",
	"http.dnsbl.sorbs.net",
	"misc.dnsbl.sorbs.net",
	"smtp.dnsbl.sorbs.net",
	"socks.dnsbl.sorbs.net",
	"spam.dnsbl.sorbs.net",
	"recent.spam.dnsbl.sorbs.net",
	"new.spam.dnsbl.sorbs.net",
	"old.spam.dnsbl.sorbs.net",
	"web.dnsbl.sorbs.net",
}

// SpamratsIPBlacklists lists to check IP against databases
var SpamratsIPBlacklists = []string{
	"all.spamrats.com",
	"auth.spamrats.com",
	"dyna.spamrats.com",
	"noptr.spamrats.com",
	"spam.spamrats.com",
}
