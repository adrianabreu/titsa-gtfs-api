import requests
import json
import xmltodict
from datetime import datetime, timedelta
import os

def extract_titsa_data(titsa_token, stop_id): 
    titsa_url = f"http://apps.titsa.com/apps/apps_sae_llegadas_parada.asp?idApp={titsa_token}&idParada={stop_id}"
    response = requests.get(titsa_url)
    ARRIVALS_KEY = "llegadas"
    if response.status_code == 200:
        response = xmltodict.parse(response.content)
        if ARRIVALS_KEY not in response or response[ARRIVALS_KEY] is None:
            return (200, None)
        else:
            return (200, response[ARRIVALS_KEY])
    else:
        return (response.status_code, None)

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

def log_call(token, arrivals):
    TINYBIRD_URL = "https://api.tinybird.co/v0/events?name=titsa_api-status"
    data = { 
        "ts": datetime.now().strftime("%Y-%m-%dT%H:%M:%S.%f"),
        "status_code": arrivals[0] ,
        "empty": 1 if arrivals[1] is None else 1
    } 
    return requests.post(url=TINYBIRD_URL, data=json.dumps(data),headers= {"Authorization": f"Bearer {token}"})

if __name__ == '__main__':
    TITSA_TOKEN = os.getenv('TITSA_TOKEN')
    STOP_IDS = [1918, 2393, 1195, 2625, 9181]
    TINYBIRD_TOKEN = os.getenv('TINYBIRD_TOKEN')

    for stop_id in STOP_IDS:
        response = extract_titsa_data(TITSA_TOKEN, stop_id)
        log_call(TINYBIRD_TOKEN, response)

        arrivals = response[1]
        if (arrivals is not None): 
            print(arrivals)
            arrivals = arrivals["llegada"]
            arrivals = arrivals if (isinstance(arrivals, list)) else [arrivals]
            for arrival in arrivals:
                parsed_arrival = parse_arrival_into_json(arrival)
                append_response = publish_data(TINYBIRD_TOKEN, parsed_arrival)