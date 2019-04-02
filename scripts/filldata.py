import csv
import requests

url="http://localhost:3000/api/v1/addbook"
headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
data = []
# r = requests.post(url, json=data, headers=headers)

with open('Henrys_Little_Library.csv') as csv_file:
    csv_reader = csv.reader(csv_file, delimiter=',')
    line_count = 0
    for row in csv_reader:
        # print(row)
        line_count += 1
        data.append({'title': row[0], 'author': row[1], 'binding': row[2], 'source': row[3]})
    print(f'Processed {line_count} lines.')

r = requests.post(url, json=data, headers=headers)