package main

import (
	"encoding/json"
	"fmt"

	"github.com/gabrielpetry/update_ip/config"
	"github.com/gabrielpetry/update_ip/hosts"
	"github.com/gabrielpetry/update_ip/iface"
	"github.com/gabrielpetry/update_ip/providers"
)

const APPNAME = "update_ip"

func main() {
	fmt.Print("\033[H\033[2J") // clear terminal

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
