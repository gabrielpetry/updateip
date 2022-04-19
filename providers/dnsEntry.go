package providers

import (
	"encoding/json"
	"fmt"
)

type DnsEntry struct {
	Host    string
	Entry   string
	Type    string
	Target  string
	Id      string
	Ttl     int
	Proxied bool
}

func (d *DnsEntry) PrintJson(entries []DnsEntry) string {
	c, _ := json.MarshalIndent(entries, "", "  ")
	return string(c)
}

func (d *DnsEntry) PrintBash(entries []DnsEntry) string {
	var bash string
	for _, entry := range entries {
		if entry.Entry != "" {
			entry.Entry = entry.Entry + "."
		}
		proxied := ""
		if entry.Proxied {
			proxied = "proxied by cloudflare"
		}
		bash += fmt.Sprintf(
			"%s\t\t%s%s # %v\n",
			entry.Target,
			entry.Entry,
			entry.Host,
			proxied)
	}

	return bash
}
