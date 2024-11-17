package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const filePath = "RLEB/heartbeat.txt"
const noResponseThresholdMinutes = 10
const pulseDeadThresholdSeconds = 1000
var logs []string

func restart() {
	// Overwrite the heartbeat to 0 to prevent cycles.
	err := ioutil.WriteFile(filePath, []byte("0\n"), 0644)
	if err != nil {
		fmt.Printf("Failed to write file: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Restart the device
	if err := exec.Command("sudo", "reboot").Run(); err != nil {
		log.Printf("Failed to reboot: %v", err)
	}
}

func main() {
	for {
		// Check if the heartbeat file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("File does not exist: %s\n", filePath)
			time.Sleep(10 * 60 * time.Second)
			continue
		}

		// Get file modification time
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Failed to get file info: %v", err)
		}
		lastModified := fileInfo.ModTime()

		// If the file hasn't been written in 10m, restart.
		timeDiff := time.Since(lastModified)
		fmt.Printf("Last heartbeat received %s ago", timeDiff)
		if timeDiff > noResponseThresholdMinutes * time.Minute {
			fmt.Printf("Restarting due to no heartbeat received in %d minutes.", noResponseThresholdMinutes)
			restart()
		}

		// Read the contents of the text file
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Failed to read file: %v", err)
		}

		// Read only the first line of the text file, cast to integer.
		line := strings.TrimSpace(string(content))
		heartbeat, err := strconv.Atoi(line)
		fmt.Printf("Reported heartbeat: %d", heartbeat)
		if err != nil {
			fmt.Printf("Failed to read heartbeat: %s\n", line)
			time.Sleep(60 * time.Second)
			continue
		}

		// If heartbeat is greater than 1000s, restart machine.
		if heartbeat > pulseDeadThresholdSeconds {
			fmt.Printf("Restarting due to heartbeat > %d seconds.", pulseDeadThresholdSeconds)
			restart()
		}

		log.Print("Monitoring hearbeat...")
		time.Sleep(5*time.Minute)
	}
}
