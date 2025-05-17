import os
from threading import Thread

import requests
import json

for existing_file in os.listdir("ui/flags"):
    os.remove("ui/flags/" + existing_file)

data = requests.get("https://api.nordvpn.com/v1/servers?limit=200000")

json_data = json.loads(data.content)

countries = {}

for entry in json_data:
    country = entry["locations"][0]["country"]
    city = country["city"]

    country["city"] = {}
    country["city"][city["id"]] = city

    if country["id"] not in countries:
        countries[country["id"]] = country
    else:
        if city["id"] not in countries[country["id"]]["city"]:
            countries[country["id"]]["city"][city["id"]] = city

def download_flag(country_code):
    print("Downloading: ", country_code)
    flag_data = requests.get("https://hatscripts.github.io/circle-flags/flags/" + str(country_code).lower() + ".svg", stream=True)
    file = open("ui/flags/" + country_code + ".svg", "wb")
    for chunk in flag_data:
        file.write(chunk)
    file.close()


for country in countries:
    Thread(target=download_flag, args=(countries[country]["code"],)).start()
