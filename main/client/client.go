package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var nodeAddressA = "192.168.56.106"
var nodeAddressB = "192.168.56.111"

func main() {

	initializeDemo()

	for true {
		input := presentationOutput()
		switch input {
		case "1":
			incrementNumber()
		case "2":
			requestMigration()
		case "3":
			queryResult()
		case "4":
			return
		default:
			fmt.Println("The provided input is invalid.")
		}
	}
}

func initializeDemo() {
	//Send B address to node A
	requestJSON, _ := json.Marshal(nodeAddressB)
	response, err := http.Post("http://"+nodeAddressA+":8080/acquireNodeIp", "application/json", bytes.NewBuffer(requestJSON))
	errorRespCheck(err, "Failed to contact the remote node.", response.Status)
}

func incrementNumber() {
	fmt.Print("Insert the number: ")
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	requestJSON, _ := json.Marshal(input[:len(input)-1])
	response, err := http.Post("http://"+nodeAddressA+":8080/increment", "application/json", bytes.NewBuffer(requestJSON))
	errorRespCheck(err, "Failed to send the request.", response.Status)
}

func requestMigration() {
	response, err := http.Post("http://"+nodeAddressA+":8080/migrate", "application/json", nil)
	errorRespCheck(err, "Failed to send the request.", response.Status)
}

func queryResult() {
	response, err := http.Post("http://"+nodeAddressB+":8080/query", "application/json", nil)
	errorRespCheck(err, "Failed to send the request.", response.Status)
	result, _ := ioutil.ReadAll(response.Body)
	fmt.Println("The result is = " + string(result))
}

func presentationOutput() string {
	fmt.Println("\nSelect an action:\n\t1 - Insert a number to increment\n\t2 - Request migration\n\t3 - Query result\n\t4 - Leave")
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return input[:1]
}

func getIpAddress(containerName string) string {
	out, err := exec.Command("bash", "-c", "podman inspect "+containerName+" | grep IPAddress").Output()
	errorCheck(err)
	return strings.Split(string(out), "\"")[3]
}

func errorRespCheck(err error, errorMsg string, respMsg string) {
	if err != nil {
		fmt.Println(errorMsg)
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Println("Request sent. Response status: ", respMsg)
}

func errorCheck(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
