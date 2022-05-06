package main

import (
	"encoding/json"
	"fmt"

	"github.com/gabrielpetry/updateip/config"
	"github.com/gabrielpetry/updateip/hosts"
	"github.com/gabrielpetry/updateip/iface"
	"github.com/gabrielpetry/updateip/lockfile"
	"github.com/gabrielpetry/updateip/providers"
)

const APPNAME = "updateip"

var conf = config.GetInstance()

func main() {
	lockfile.Lock()
	defer lockfile.Unlock()

	c := providers.Cloudflare{}
	c.New(
		conf.Provider.Cloudflare.APIKey,
		conf.Provider.Cloudflare.APIEmail,
		conf.Provider.Cloudflare.Hostname)

	if !conf.Readonly {
		fmt.Println("updating cloudflare entries")
		Cloudflare(&c)
	}

	fmt.Println("updating hosts file")
	c.GetDnsEntries()
	err := hosts.Save(c.DnsEntries)
	if err != nil {
		fmt.Println(err)
	}
}

func PrintJson(v interface{}) {
	c, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(c))
}

func Cloudflare(c *providers.Cloudflare) {
	fmt.Println("updating cloudflare entries")
	iff := iface.Iface{}
	ifaces := []iface.Iface{}

	if conf.Ifaces.Local {
		localIff, _ := iff.LocalAddresses()
		ifaces = append(ifaces, localIff...)
	}

	if conf.Ifaces.External {
		external, _ := iff.ExternalAddress()
		ifaces = append(ifaces, external)
	}

	if len(ifaces) == 0 {
		fmt.Println("no interfaces to update")
		return
	}

	dnsEntries := c.IfaceToDnsEntry(ifaces)

	for _, entry := range dnsEntries {
		err := c.CreateOrUpdateEntry(&entry)
		if err != nil {
			fmt.Println("error creating entry:", err)
		}
	}
}
