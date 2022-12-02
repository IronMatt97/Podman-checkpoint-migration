from http.server import BaseHTTPRequestHandler, HTTPServer
from multiprocessing import Process
import os
import requests
import time
import json
import sys

hostName = "0.0.0.0"
fallbackNode_fileName = "fallbacknodes.txt"

class Unbuffered(object):
    '''
    'Unbuffered mode' allows for log messages to be immediately dumped to the stream instead of being buffered.
    '''
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
        message = post_data.decode('utf-8')
        number=int(message[1:len(message)-1])
        print("The number to increment is ",number)
        
        time.sleep(3)
        result = number+1
        print("The result is ready : ",result,"\n\tSending the response ...")

        try:
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(bytes(json.dumps(result), "utf-8"))
        except ConnectionResetError:
            print("Seems like this container has been migrated. The occurred exception is ",sys.exc_info()[0])
            self.close_connection
            
            # Acquire the new node address from the local file containing fallbackNode IP
            with open(fallbackNode_fileName, 'r') as f:
                fallbackNode = f.readline()
                print("Fallback node address acquired: ",fallbackNode)

            # Send the result to the new node    
            payload = str(result)
            requests.post('http://'+ fallbackNode +':8080/receiveMigrationRes', json = payload)
        print("Response sent.")

class MigrationListener(BaseHTTPRequestHandler):

    def do_POST(self):
        print("A migration has been requested. Acquiring the fallback node address before the checkpoint...")
        content_length = int(self.headers['Content-Length']) 
        post_data = self.rfile.read(content_length) 
        message = post_data.decode('utf-8')
        fallbackNode=message[1:len(message)-1]
        print("Fallback node acquired: ",fallbackNode)
        
        #Write the address to a local file
        with open(fallbackNode_fileName, 'w') as f:
                f.write(fallbackNode)
                print("Fallback node address stored.")
        
        self.send_response(200)
        self.send_header("Content-type", "application/json")
        self.end_headers()   

def serve(webServer):
    webServer.serve_forever()

if __name__ == "__main__":
    sys.stdout = Unbuffered(sys.stdout) # Use unbuffered output
    
    # Prepare the handlers
    executor = HTTPServer((hostName, 8080), Executor)
    migrationListener = HTTPServer((hostName, 8081), MigrationListener)

    # Start the listeners on different ports, using different processes
    Process(target=serve, args=[executor]).start()
    Process(target=serve, args=[migrationListener]).start()
    print("Container services correctly initialized.")

    # Serve until a KeyboardInterrupt occurs
    try:
        while True:
            time.sleep(60)    
    except KeyboardInterrupt:
        executor.server_close()
        migrationListener.server_close()
        pass
    
    print("\nServices closed. The container will close.\n")
    os._exit(1)