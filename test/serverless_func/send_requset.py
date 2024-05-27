import requests
import json

file_path = "./testimg.png"
with open(file_path, "rb") as file:
    file_content = file.read()

val = file_content.decode('utf-8')

data = {
    "username": "testuser",
    "image": val,
    "x": 20,
    "y": 20,
    "r": 255,
    "g": 255,
    "b": 255,
    "mark": "water_mark",
    "mysqlIp": "10.2.1.25",
    "rate": 0.8,
    "threshold1": 50,
    "threshold2": 150,
    "kernel_x": 3,
    "kernel_y": 3,
    "status": "start"
}

url = "http://127.0.0.1:8080"
headers = {'Content-Type': 'application/json'}
response = requests.post(url, headers=headers, data=json.dumps(data))

print(response.status_code)
print(response.text)