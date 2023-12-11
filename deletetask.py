# Poll the API every X amount of time. If the task is completed, move to completed task table and delete from current task table.
import requests
import json
import time
import datetime

base_url = "http://10.0.0.5:8080/tasks/"

r = requests.get(base_url)
r = r.json()

completed = [(task['id'], task['completed']) for task in r]
for val in completed: 
    base_url = "http://10.0.0.5:8080/completed/"
    if val[1] == True:
        r = requests.post(f'{base_url}{val[0]}')
        print(r)
        base_url = "http://10.0.0.5:8080/delete/"
        r = requests.delete(f'{base_url}{val[0]}')
        print(f"Task {val[0]} is completed and has been moved to the completed task table.")
    else:
        print(f"Task {val[0]} is not completed yet.")
