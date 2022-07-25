import requests
import json
import xmltodict
from datetime import datetime, timedelta
import os

if __name__ == '__main__':
    titsa_token = os.getenv('TITSA_TOKEN')
    stop_id = 1918
    titsa_url = f"http://apps.titsa.com/apps/apps_sae_llegadas_parada.asp?idApp={titsa_token}&idParada={stop_id}"
    tinybird_url = "https://api.tinybird.co/v0/events?name=realtime"
    tinybird_token = os.getenv('TINYBIRD_TOKEN')

    response = requests.get(titsa_url)

    if response.status_code == 200:
        response = xmltodict.parse(response.content)
        if "llegadas" not in response:
            sys.exit(f"Empty Titsa API answer")
        if response["llegadas"] is None:
            sys.exit(f"Empty Titsa API answer")

        for arrival in response["llegadas"]:
            body = response["llegadas"][arrival]
            mins_next_arrival = response["llegadas"][arrival]["minutosParaLlegar"]
            arrival_time = datetime.strptime(body["hora"], '%d/%m/%Y %H:%M:%S') + timedelta(minutes=int(mins_next_arrival))
            append_response = requests.post(url=tinybird_url, data=json.dumps({
                "line": str(body["linea"]),
                "stop_id": body["codigoParada"],
                "calendar_date": arrival_time.strftime("%Y%m%d"),
                "arrival_time": arrival_time.strftime("%H:%M:%S")
            }),headers= {"Authorization": f"Bearer {tinybird_token}"})
            print(f"Data appended {append_response.status_code}")
    else:
        sys.exit(f"Titsa API unavailable {response.status_code}")