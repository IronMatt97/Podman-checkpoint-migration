from http.server import BaseHTTPRequestHandler, HTTPServer
import time
import json
import sys

hostName = "0.0.0.0"
serverPort = 8080

class Unbuffered(object):
   def __init__(self, stream):
       self.stream = stream
   def write(self, data):
       self.stream.write(data)
       self.stream.flush()
   def writelines(self, datas):
       self.stream.writelines(datas)
       self.stream.flush()
   def __getattr__(self, attr):
       return getattr(self.stream, attr)

class Executor(BaseHTTPRequestHandler):

    def do_POST(self):
        content_length = int(self.headers['Content-Length']) 
        post_data = self.rfile.read(content_length) 
        number = int((post_data.decode('utf-8'))[1:2])
        print("Before wait")
        time.sleep(3)
        print("After wait")
        result = number+1
        print("Incremented variable")
        try:
            self.send_response(200)
            print("Response sent")
            self.send_header("Content-type", "application/json")
            print("Header sent")
            self.end_headers()
            print("Headers end sent")
            self.wfile.write(bytes(json.dumps(result), "utf-8"))
            print("POST completed.")
        except:
            print("Oops!", sys.exc_info()[0], "occurred.")
            #TODO: reconnect to the new node manually to communicate the result.
            #Exception class: <class 'ConnectionResetError'>

if __name__ == "__main__":
    sys.stdout = Unbuffered(sys.stdout)
    webServer = HTTPServer((hostName, serverPort), Executor)
    print("Executor server started http://%s:%s" % (hostName, serverPort))

    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass

    webServer.server_close()
    print("\nServer stopped.\n")