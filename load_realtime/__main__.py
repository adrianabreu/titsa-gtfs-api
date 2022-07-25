import requests
import json
import xmltodict
from datetime import datetime, timedelta
import os
import sys

def extract_titsa_data(titsa_token, stop_id): 
    print(titsa_token)
    titsa_url = f"http://apps.titsa.com/apps/apps_sae_llegadas_parada.asp?idApp={titsa_token}&idParada={stop_id}"
    response = requests.get(titsa_url)
    ARRIVALS_KEY = "llegadas"
    if response.status_code == 200:
        response = xmltodict.parse(response.content)
        if ARRIVALS_KEY not in response or response[ARRIVALS_KEY] is None:
            sys.exit(f"Empty Titsa API answer")
        else:
            print(response)
            return response[ARRIVALS_KEY]
    else:
        sys.exit(f"Titsa API unavailable {response.status_code}")

def parse_arrival_into_json(arrival):
    mins_next_arrival = arrival["minutosParaLlegar"]
    arrival_time = datetime.strptime(arrival["hora"], '%d/%m/%Y %H:%M:%S') + timedelta(minutes=int(mins_next_arrival))
    return {
        "line": int(arrival["linea"]),
        "stop_id": int(arrival["codigoParada"]),
        "calendar_date": arrival_time.strftime("%Y%m%d"),
        "arrival_time": arrival_time.strftime("%H:%M:%S")
    }

def publish_data(token, data):
    TINYBIRD_URL = "https://api.tinybird.co/v0/events?name=realtime"
    return requests.post(url=TINYBIRD_URL, data=json.dumps(data),headers= {"Authorization": f"Bearer {token}"})

if __name__ == '__main__':
    TITSA_TOKEN = os.getenv('TITSA_TOKEN')
    STOP_ID = 1918
    TINYBIRD_TOKEN = os.getenv('TINYBIRD_TOKEN') 
    arrivals = extract_titsa_data(TITSA_TOKEN, STOP_ID)
    
    for arrival in arrivals:
        parsed_arrival = parse_arrival_into_json(arrivals[arrival])
        append_response = publish_data(TINYBIRD_TOKEN, parsed_arrival) 
        print(f"Data appended {append_response.status_code}")

