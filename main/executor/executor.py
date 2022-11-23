from http.server import BaseHTTPRequestHandler, HTTPServer
import requests
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
        print("A request has been received.")
        content_length = int(self.headers['Content-Length']) 
        post_data = self.rfile.read(content_length) 
        message = post_data.decode('utf-8').split("|")
        fallbackNode=message[0][1:len(message[0])]  
        number=int(message[1][:len(message[1])-1])
        print("FallbackNode acquired: ",fallbackNode,"\nNumber to increment= ",number)
        time.sleep(3)
        result = number+1
        print("Result: ",result)
        try:
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(bytes(json.dumps(result), "utf-8"))
            print("Response sent.")
        except:
            #except ConnectionResetError
            print(sys.exc_info()[0], "has occurred: a migration just happened.")
            self.close_connection
            payload = str(result)
            requests.post('http://'+ fallbackNode +':8080/receiveMigrationRes', json = payload)
            print("Response sent.")
           

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
    