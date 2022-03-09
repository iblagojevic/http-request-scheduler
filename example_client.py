import requests
import concurrent
from concurrent.futures import ThreadPoolExecutor
import random

rng = range(1, 100)
service_url = 'http://localhost:9292/'
threads = 10


def make_request(o_id, o_type):
    body = {
        "action": "POST",
        "url": "http://your-local-endpoint/api/submit",
        "payload": {
            "id": o_id,
            "type": o_type
        },
        "delay": random.randint(1, 100)
    }
    r = requests.post(service_url, json=body)
    return r.text

with ThreadPoolExecutor(max_workers=threads) as executor:
    future_to_url = {executor.submit(make_request, r, f"Event_{r}") for r in rng}
    for future in concurrent.futures.as_completed(future_to_url):
        try:
            response = future.result()
            if response == "accepted":
                print("Request is submitted for scheduling.")
            else:
                pass
        except Exception as e:
            print(f'Something went wrong:{str(e)}')

