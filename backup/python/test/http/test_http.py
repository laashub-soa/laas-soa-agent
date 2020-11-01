import requests

if __name__ == '__main__':
    r = requests.post("http://0.0.0.0:8080/agent/consume_thing", {}, {})
    resp = r.json()
    print(resp)
