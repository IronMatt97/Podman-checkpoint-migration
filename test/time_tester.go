package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var checkpointArchiveName = "test.tar.gz"
var imageName = "localhost/counter"
var containerName = "test"
var iterations = 100

func main() {
	csvFile, _ := os.Create("infoData.csv")
	csvwriter := csv.NewWriter(csvFile)
	csvwriter.Write([]string{"checkpoint time", "restore time", "checkpoint+restore time", "archive size (Bytes)", "time/dim ratio (microseconds/Bytes)"})

	for i := 1; i <= iterations; i++ {
		fmt.Println("Iteration ", i, "/", iterations)
		cleanupEnvironment()
		start()
		c_time := checkpoint()
		r_time := restore()
		tot_time := c_time + r_time
		archive_size := checkpointArchiveSize()
		ratio := float64(float64(tot_time.Microseconds()) / float64(archive_size))
		cleanupEnvironment()
		csvwriter.Write([]string{c_time.String(), r_time.String(), tot_time.String(), strconv.FormatInt(archive_size, 10), strconv.FormatFloat(ratio, 'f', 5, 32)})
	}
	csvwriter.Flush()
	csvFile.Close()
}

func checkpoint() time.Duration {
	startTime := time.Now()
	exec.Command("podman", "container", "checkpoint", containerName, "-e", checkpointArchiveName).Run()
	return time.Since(startTime)
}

func restore() time.Duration {
	startTime := time.Now()
	exec.Command("podman", "container", "restore", "-i", checkpointArchiveName).Run()
	return time.Since(startTime)
}

func start() {
	exec.Command("podman", "run", "-dt", "--name="+containerName, imageName).Run()
}

func checkpointArchiveSize() int64 {
	checkpointFile, _ := os.Stat(checkpointArchiveName)
	return checkpointFile.Size()
}

func cleanupEnvironment() {
	exec.Command("podman", "stop", containerName, "-t", "0").Run()
	exec.Command("podman", "rm", containerName).Run()
}
