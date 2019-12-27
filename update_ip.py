#!/bin/python

#
import requests, json, os
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
    last_external_ip  = ""
    headers           = {}
    entry_id          = ""


    def __init__(self, token, email, entry, domain, zone_id, logfile):
        self.cloudflare_token = token
        self.cloudflare_email = email
        self.cloudflare_entry = entry
        self.cloudflare_domain = domain
        self.cloudflare_dns_entry = entry + "." + domain
        self.cloudflare_api_base += zone_id
        self.cloudflare_zone_id = zone_id

        self.external_ip = self.get_external_ip()

    def get_external_ip(self):
        return requests.get('http://ifconfig.me').text

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
            "ttl": 120,
            "proxied": False
            })


    def update_entry(self):
        return self.request('put', '/dns_records/' + self.entry_id, {
            "type": "A",
            "name": self.cloudflare_dns_entry,
            "content": self.external_ip,
            "ttl": 120,
            "proxied": False
        })

    def change_dns(self):
        entry_exists = self.check_if_entry_exists()

        if entry_exists:
            self.entry_id = entry_exists['id']
            return self.update_entry()

        return self.create_entry()



if __name__ == "__main__":

    load_dotenv()

    cloudflare = update_ip(
        os.getenv('cloudflare_api'),
        os.getenv('cloudflare_email'),
        os.getenv('cloudflare_entry'),
        os.getenv('cloudflare_domain'),
        os.getenv('cloudflare_zone_id'),
        os.getenv('logpath')
    )

    print(cloudflare.change_dns().json())

    # print(cloudflare.get_external_ip())
    # print(cloudflare.entryExists())
    # print(cloudflare.check_if_entry_exists())
    # cloudflare.create_entry();
