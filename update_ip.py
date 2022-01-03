#!/bin/python
import requests, json, os
import netifaces as ni
from dotenv import load_dotenv


class update_ip:
    cloudflare_token  = ""
    cloudflare_email  = ""
    cloudflare_entry  = ""
    cloudflare_domain = ""
    cloudflare_dns_entry = ""
    cloudflare_zone_id = ""
    cloudflare_api_base = "https://api.cloudflare.com/client/v4/zones/"
    logfile           = ""
    external_ip       = ""
    headers           = {}
    entry_id          = ""
    ttl               = 120
    proxied           = False
    interface         = "external"

    def __init__(self,
                token,
                email,
                entry,
                domain,
                zone_id,
                logfile,
                interface,
                ttl,
                proxied):
        self.cloudflare_token = token
        self.cloudflare_email = email
        self.cloudflare_entry = entry
        self.cloudflare_domain = domain
        self.cloudflare_dns_entry = entry + "." + domain
        self.cloudflare_api_base += zone_id
        self.cloudflare_zone_id = zone_id
        self.ttl = ttl or 120
        self.interface = interface or "external"

        if proxied and proxied.lower() == "true":
            self.proxied = True
        else:
            self.proxied = False

        if interface is None or interface == 'external':
            self.external_ip = self.get_external_ip()
        else:
            self.external_ip = self.get_interface_ip(interface)

    def get_external_ip(self):
        return requests.get('http://ifconfig.me').text

    def get_interface_ip(self, interface):
        print(ni.ifaddresses(interface))
        return ni.ifaddresses(interface)[ni.AF_INET][0]['addr']


    def request(self, method, endpoint, data={}):
        headers = {
            "Accept": "application/json",
            "Content-Type": "application/json",
            "X-Auth-Key": self.cloudflare_token,
            "X-Auth-Email": self.cloudflare_email
        }

        return requests.request(
            method,
            self.cloudflare_api_base + endpoint,
            data=json.dumps(data),
            headers=headers
        )

    def check_if_entry_exists(self):
        entries = self.request('get', '/dns_records/').json()
        entry   = list(filter(
            lambda x: x['name'] == self.cloudflare_dns_entry,
            entries['result']))

        if entry and entry[0]:
            return entry[0]

        return False

    def create_entry(self):
        return self.request('post', '/dns_records', {
            "type": "A",
            "name": self.cloudflare_dns_entry,
            "content": self.external_ip,
            "ttl": self.ttl,
            "proxied": self.proxied
            })


    def update_entry(self):
        return self.request('put', '/dns_records/' + self.entry_id, {
            "type": "A",
            "name": self.cloudflare_dns_entry,
            "content": self.external_ip,
            "ttl": self.ttl,
            "proxied": self.proxied
        })

    def change_dns(self):
        entry_exists = self.check_if_entry_exists()
        print("Changing DNS entry for " + self.cloudflare_dns_entry)
        print("Current IP: " + self.external_ip)
        print("Interface IP: " + self.interface)


        if entry_exists:
            self.entry_id = entry_exists['id']
            return self.update_entry()

        return self.create_entry()



if __name__ == "__main__":

    load_dotenv()

    for entry in os.getenv('cloudflare_entry').split(','):
        print("entry: " + entry)
        print(os.getenv(entry + "_interface"),)
        cloudflare = update_ip(
            token=os.getenv('cloudflare_api'),
            email=os.getenv('cloudflare_email'),
            entry=entry,
            domain=os.getenv('cloudflare_domain'),
            zone_id=os.getenv('cloudflare_zone_id'),
            logfile=os.getenv('logpath'),
            interface=os.getenv(entry + "_interface"),
            ttl=os.getenv(entry + "_ttl"),
            proxied=os.getenv(entry + "_proxied")
        )

        print(cloudflare.change_dns().json())

