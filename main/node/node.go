package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var checkpointArchiveName = "checkpoint.tar.gz"
var checkpointArchiveSizeLimit = 10 * 1024
var containerAddress = ""
var otherNodeIP = ""
var resultNumber = ""
var requestNumber = ""

func main() {
	if spawnExecutor() {
		containerAddress = initializeExecutor() //If '-executor' flag is specified, a local executor is spawned
	}
	fmt.Println("Node initialized.")
	http.HandleFunc("/acquireNodeIp", acquireIp)
	http.HandleFunc("/increment", increment)
	http.HandleFunc("/migrate", migrate)
	http.HandleFunc("/query", getResult)
	http.HandleFunc("/restore", completeMigration)
	http.HandleFunc("/receiveMigrationRes", receiveResult)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Server initialization error.\n")
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

func initializeExecutor() string {
	cleanupEnvironment()
	err := exec.Command("podman", "run", "-dt", "--name=executor", "localhost/executor").Run()
	errorCheck(err)
	ip := getIpAddress("executor")
	fmt.Println("Remote executor acquired. IP: ", ip)
	return ip
}

func acquireIp(w http.ResponseWriter, r *http.Request) {
	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal([]byte(requestBody), &otherNodeIP)
	fmt.Println("Remote node ip acquired -> ", otherNodeIP)
}

func increment(w http.ResponseWriter, r *http.Request) {
	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal([]byte(requestBody), &requestNumber)
	fmt.Println("A number has been received: ", requestNumber)
	go submitAsyncRequest()
	fmt.Println("The number has been submitted.")
}

func migrate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("A migration on the remote node has been requested.")

	// First of all, contact the container (port 8081) to send the fallback node address
	requestJSON, _ := json.Marshal(otherNodeIP)
	response, err := http.Post("http://"+containerAddress+":8081", "application/json", bytes.NewBuffer(requestJSON))
	errorRespCheck(err, "Failed to send request", response.Status)

	// Checkpoint the container
	err = exec.Command("podman", "container", "checkpoint", "executor", "-e", checkpointArchiveName, "--tcp-established").Run()
	errorCheck(err)
	fmt.Println("\t...Container checkpointed.")

	// Prepare a request body to send the checkpoint .tar file
	fileDir, _ := os.Getwd() // Get current path
	filePath := path.Join(fileDir, checkpointArchiveName)

	file, _ := os.Open(filePath) // Open file
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(checkpointArchiveName, filepath.Base(file.Name()))
	io.Copy(part, file) // Copy file bytes in a multipart form data file
	writer.Close()

	// Send the checkpoint .tar file to the remote node
	r, _ = http.NewRequest("POST", "http://"+otherNodeIP+":8080/restore", body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	client.Do(r)
	fmt.Println("\t...Checkpoint sent to ", otherNodeIP)
}

func completeMigration(w http.ResponseWriter, r *http.Request) {
	fmt.Println("A checkpoint file has been received.")

	// Receive the checkpoint .tar file
	r.ParseMultipartForm(int64(checkpointArchiveSizeLimit))
	file, handler, err := r.FormFile(checkpointArchiveName) // Get the form file
	errorCheck(err)
	defer file.Close()

	fmt.Printf("Uploaded file specs:\nName -> %+v\nSize -> %+v\nMIME Header -> %+v\n", handler.Filename, handler.Size, handler.Header)
	currDir, _ := os.Getwd()
	tempFile, err := ioutil.TempFile(currDir, "checkpoint-*.tar.gz") // Prepare the temporary file
	errorCheck(err)
	defer tempFile.Close()

	fileBytes, _ := ioutil.ReadAll(file) // Read file content in a byte array
	tempFile.Write(fileBytes)            // Write the byte array in the temporary file
	fmt.Printf("Checkpoint file %s successfully received.\n", tempFile.Name())

	// Restore the execution
	restoreExecution(tempFile.Name())
}

func restoreExecution(fileName string) {
	err := exec.Command("podman", "container", "restore", "-i", fileName, "--tcp-established").Run()
	errorCheck(err)
	fmt.Println("Container restored. Migration completed.")
}

func receiveResult(w http.ResponseWriter, r *http.Request) {
	fmt.Println("A migrated container just sent a result.")
	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal([]byte(requestBody), &resultNumber)
	fmt.Println("The result is ready: " + resultNumber)
}

func getResult(w http.ResponseWriter, r *http.Request) {
	fmt.Println("A result has been queried.")
	json.NewEncoder(w).Encode(string(resultNumber))
}

func submitAsyncRequest() {
	// Send the number to actually increment
	requestJSON, _ := json.Marshal(requestNumber)
	response, err := http.Post("http://"+containerAddress+":8080", "application/json", bytes.NewBuffer(requestJSON))
	errorRespCheck(err, "Failed to send request", response.Status)
	result, _ := ioutil.ReadAll(response.Body)
	resultNumber = string(result)
	fmt.Println("The result is ready: " + resultNumber)
}

func getIpAddress(containerName string) string {
	out, err := exec.Command("bash", "-c", "podman inspect "+containerName+" | grep IPAddress").Output()
	errorCheck(err)
	return strings.Split(string(out), "\"")[3]
}

func cleanupEnvironment() {
	exec.Command("podman", "stop", "executor").Run()
	exec.Command("podman", "rm", "executor").Run()
}

func spawnExecutor() bool {
	if len(os.Args) > 1 {
		return os.Args[1] == "-executor"
	}
	return false
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
