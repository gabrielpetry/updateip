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
	iface := iface.Iface{}
	ifaces, _ := iface.LocalAddresses()

	external, _ := iface.ExternalAddress()
	ifaces = append(ifaces, external)

	dnsEntries := c.IfaceToDnsEntry(ifaces)

	for _, entry := range dnsEntries {
		err := c.CreateOrUpdateEntry(&entry)
		if err != nil {
			fmt.Println("error creating entry:", err)
		}
	}
}
