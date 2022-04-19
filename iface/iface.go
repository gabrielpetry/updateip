package iface

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gabrielpetry/update_ip/config"
)

type Iface struct {
	Name string
	Addr string
}

var hostname = ""

func getHostname() string {
	if hostname == "" {
		hostname, _ = os.Hostname()
		hostname = strings.ToLower(hostname)
		hostname = strings.ReplaceAll(hostname, ".local", "")
		hostname = strings.ReplaceAll(hostname, ".", "")
		hostname = strings.ReplaceAll(hostname, "-", "")
		hostname = strings.ReplaceAll(hostname, "_", "")
	}
	return hostname
}

func (i *Iface) ExternalAddress() (entry Iface, err error) {
	iface := Iface{}
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		resp, err = http.Get("https://www.giot.ir/webservices/returnmyip.php")
		if err != nil {
			resp, err = http.Get("https://myip.dnsomatic.com/")
			if err != nil {
				return iface, nil
			}
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	iface.Name = getHostname()
	iface.Addr = string(body)
	return iface, nil
}

func (i *Iface) LocalAddresses() (entries []Iface, err error) {
	config := config.GetInstance()

	ifaces := []Iface{}

	systemIfaces, err := net.Interfaces()
	if err != nil {
		return ifaces, err
	}

	for _, i := range systemIfaces {
		addrs, err := i.Addrs()

		// check if it`s a valid name based on a regex
		match, _ := regexp.MatchString(config.Ifaces.Regex.Name, i.Name)
		if config.Ifaces.Regex.Name != "" && !match {
			continue
		}

		if err != nil {
			// return ifaces, nil
			fmt.Println(fmt.Errorf("error getting interface %s addresses: %w, ain't fatal", i.Name, err))
			continue
		}

		for _, addr := range addrs {
			ip := strings.Split(addr.String(), "/")[0] // 192.168.0.22/24
			match, _ = regexp.MatchString(config.Ifaces.Regex.Addr, ip)
			if !match {
				continue
			}

			// will be store as: macbook.en0.example.com
			name := fmt.Sprintf("%s.%s", getHostname(), i.Name)

			ifaces = append(ifaces, Iface{
				Name: name,
				Addr: ip})
		}
	}

	// usually en0 or eth0 will be added first, before tun or wlan,
	// so we will return the first one as a .local, maybe a dump idea
	ifaces = append(ifaces, Iface{
		Name: getHostname() + ".local",
		Addr: ifaces[0].Addr})

	return ifaces, nil
}
