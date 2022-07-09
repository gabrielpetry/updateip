package hosts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gabrielpetry/updateip/providers"
)

var startIdentifier = "##### bypass cloudflare dns #####"
var endIdentifier = "##### End bypass #####"

func filter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func Save(entries []providers.DnsEntry) error {
	hosts := []string{}

	hostsFile, err := os.Open("/etc/hosts")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(hostsFile)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		hosts = append(hosts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// remove any old key from domains
	for _, entry := range entries {
		hosts = filter(hosts, func(s string) bool {
			if strings.Contains(s, entry.Host) {
				return false
			}

			if strings.Contains(s, endIdentifier) {
				return false
			}

			if strings.Contains(s, startIdentifier) {
				return false
			}

			if  s == "" {
				return false
			}

			return true
		})
	}

	hosts = append(hosts, "\n#"+startIdentifier)

	for _, entry := range entries {
		if entry.Entry != "" {
			entry.Entry = entry.Entry + "." // append a dot to the entry
		}

		proxied := "Proxied"
		if !entry.Proxied {
			proxied = ""
		}

		hosts = append(hosts, fmt.Sprintf("%s\t\t%s%s # %v", entry.Target, entry.Entry, entry.Host, proxied))
	}

	hosts = append(hosts, endIdentifier+"\n")
	data := []byte(strings.Join(hosts, "\n"))
	fmt.Print(string(data))
	return os.WriteFile("/etc/hosts", data, 0644)
}
