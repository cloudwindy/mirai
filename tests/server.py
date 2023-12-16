import json
from http.server import HTTPServer, BaseHTTPRequestHandler

host = ('localhost', 3000)
climate = None
with open('climate.json') as f:
    climate = json.load(f)


class Request(BaseHTTPRequestHandler):
    def do_GET(self):
        name = self.path.strip('/api/climate/')
        if name == "":
            return
        self.send_response(200)
        self.send_header('Content-Type', 'text/plain')
        self.end_headers()
        temp = []
        for record in climate:
            if record['Country'] == name:
                temp.append(float(record['AverageTemperature']))
        avg = sum(temp) / len(temp)
        self.wfile.write(str(avg).encode())


if __name__ == '__main__':
    with HTTPServer(host, Request) as srv:
        print('部署完成')
        srv.serve_forever()
