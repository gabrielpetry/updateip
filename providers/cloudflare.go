package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/gabrielpetry/updateip/iface"
)

const validTypes = "A" // A,AAAA,CNAME
// var conf = config.GetInstance()

type Cloudflare struct {
	Hostname   string     `json:"Hostname"`
	DnsEntries []DnsEntry `json:"dns_entries"`
	apiKey     string
	email      string
	api        *cloudflare.API
	ctx        context.Context
	zoneId     string
}

func (c *Cloudflare) IfaceToDnsEntry(ifaces []iface.Iface) []DnsEntry {
	dnsEntries := []DnsEntry{}
	isIpv4, _ := regexp.Compile("[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}")
	isLocalAddr, _ := regexp.Compile("^(127|10|172|192)")

	for _, iface := range ifaces {
		entryType := "AAAA" // ipv4
		proxied := true

		if match := isIpv4.MatchString(iface.Addr); match {
			entryType = "A"
		}

		if match := isLocalAddr.MatchString(iface.Addr); match {
			proxied = false
		}

		dnsEntry := DnsEntry{
			Host:    c.Hostname,
			Entry:   iface.Name,
			Type:    entryType,
			Target:  iface.Addr,
			Ttl:     120,
			Proxied: proxied,
		}

		dnsEntries = append(dnsEntries, dnsEntry)
	}

	return dnsEntries
}

func (c *Cloudflare) New(apiKey string, email string, Hostname string) error {
	c.apiKey = apiKey
	c.email = email
	c.Hostname = Hostname
	c.ctx = context.Background()

	api, err := cloudflare.New(c.apiKey, c.email)
	if err != nil {
		return err
	}

	c.api = api

	zoneId, err := api.ZoneIDByName(c.Hostname)
	if err != nil {
		return err
	}

	c.zoneId = zoneId
	c.GetDnsEntries()
	return nil
}

func (c *Cloudflare) GetDnsEntries() ([]DnsEntry, error) {
	dnsEntries := []DnsEntry{}

	records, err := c.api.DNSRecords(c.ctx, c.zoneId, cloudflare.DNSRecord{})
	if err != nil {
		return dnsEntries, err
	}

	for _, record := range records {
		// check for a valid type
		if !strings.Contains(validTypes, record.Type) {
			continue
		}
		entry := strings.Replace(record.Name, "."+c.Hostname, "", 1)
		if entry == record.Name {
			entry = ""
		}

		dnsEntry := DnsEntry{
			Host:    c.Hostname,
			Entry:   entry,
			Type:    record.Type,
			Target:  record.Content,
			Id:      record.ID,
			Ttl:     record.TTL,
			Proxied: *record.Proxied,
		}

		dnsEntries = append(dnsEntries, dnsEntry)
		c.DnsEntries = dnsEntries
	}

	sort.SliceStable(dnsEntries, func(i, j int) bool {
		return dnsEntries[i].Entry < dnsEntries[j].Entry
	})

	return dnsEntries, nil
}

func (c *Cloudflare) checkEntryExists(entry *DnsEntry) string {
	for _, dnsEntry := range c.DnsEntries {
		if dnsEntry.Entry == entry.Entry && dnsEntry.Host == entry.Host {
			return dnsEntry.Id
		}
	}
	return ""
}

func (c *Cloudflare) CreateOrUpdateEntry(entry *DnsEntry) error {
	id := c.checkEntryExists(entry)
	if id == "" {
		return c.CreateEntry(entry)
	}

	return c.UpdateEntry(entry, id)
}

func (c *Cloudflare) UpdateEntry(entry *DnsEntry, id string) error {
	rr := cloudflare.DNSRecord{
		Type:    entry.Type,
		Name:    entry.Entry + "." + c.Hostname,
		Content: entry.Target,
		TTL:     entry.Ttl,
		Proxied: &entry.Proxied,
	}
	vv, _ := json.MarshalIndent(rr, "", "  ")
	fmt.Println("Updating Record: ", string(vv))

	return c.api.UpdateDNSRecord(c.ctx, c.zoneId, id, rr)
}

func (c *Cloudflare) CreateEntry(entry *DnsEntry) error {

	rr := cloudflare.DNSRecord{
		Type:    "A",
		Name:    entry.Entry + "." + c.Hostname,
		Content: entry.Target,
		Proxied: &entry.Proxied,
		TTL:     entry.Ttl,
	}

	vv, _ := json.MarshalIndent(rr, "", "  ")
	fmt.Println("Creating Record: ", string(vv))
	response, err := c.api.CreateDNSRecord(c.ctx, c.zoneId, rr)
	if err != nil {
		return err
	}
	id := response.Result.ID

	err = c.api.UpdateDNSRecord(c.ctx, c.zoneId, id, rr)
	if err != nil {
		return err
	}
	return nil
}
