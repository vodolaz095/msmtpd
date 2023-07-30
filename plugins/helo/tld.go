package helo

import (
	_ "embed"
	"strings"
)

// Data can be acquired by calling
// wget https://data.iana.org/TLD/tlds-alpha-by-domain.txt

//go:embed tlds-alpha-by-domain.txt
var rawData string

var TopListDomains []string

func init() {
	tlds := strings.Split(rawData, "\n")
	TopListDomains = make([]string, len(tlds)-1)
	for i := range tlds[1:] {
		TopListDomains[i] = strings.TrimSpace(tlds[i])
	}
}
