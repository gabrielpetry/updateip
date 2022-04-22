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

func main() {
	lockfile.Lock()
	defer lockfile.Unlock()

	config := config.GetInstance()
	iface := iface.Iface{}

	ifaces, _ := iface.LocalAddresses()
	c := providers.Cloudflare{}
	c.New(
		config.Provider.Cloudflare.APIKey,
		config.Provider.Cloudflare.APIEmail,
		config.Provider.Cloudflare.Hostname)

	external, _ := iface.ExternalAddress()
	ifaces = append(ifaces, external)

	dnsEntries := c.IfaceToDnsEntry(ifaces)

	for _, entry := range dnsEntries {
		err := c.CreateOrUpdateEntry(&entry)
		if err != nil {
			fmt.Println("error creating entry:", err)
		}
	}

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
