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
const logFilePath = "Pulse/pulse_logs.txt"
const noResponseThresholdMinutes = 10
const pulseDeadThresholdSeconds = 1000
var logs []string

func flushLog(){
	// Important: We have to not call printAndAddToLog() because it might loop.

	directory, _ := os.Getwd()
	flush_message := fmt.Sprintf("[%s] Flushing %d logs.", directory,  len(logs)+1)
	log.Print(flush_message)
	logs = append(logs, flush_message)
	// Write to pulse logs.
	existingContent, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		log.Println("Log file not found. Creating a new one.")
	}

	// Append the new message to the existing content.
	// Optimize this if we ever have much larger chunk sizes in the future.
	newContent := existingContent
	for _, logRow := range(logs){
		newContent = append(newContent, []byte(logRow)...)
	}

	// Write the updated content back to the log file
	if err := ioutil.WriteFile(logFilePath, newContent, 0644); err != nil {
		error_message := fmt.Sprintf("Failed to write to log file: %v", err)
		log.Print(error_message)
		logs = append(logs, error_message)
	}

	logs = []string{}
}

func printAndAddToLog(newLog string){
	log.Print(newLog)
	timestamped_log := fmt.Sprintf("%s - %s", time.Now().UTC().Format(time.RFC3339), newLog)
	logs = append(logs, timestamped_log)
	if len(logs) > 100{
		flushLog()
	}
}

func restart() {
	// Overwrite the heartbeat to 0 to prevent cycles.
	err := ioutil.WriteFile(filePath, []byte("0\n"), 0644)
	if err != nil {
		printAndAddToLog(fmt.Sprintf("Failed to write file: %v", err))
	}
	flushLog()
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
			printAndAddToLog(fmt.Sprintf("File does not exist: %s\n", filePath))
			time.Sleep(60 * time.Second)
			continue
		}

		// Get file modification time
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			printAndAddToLog(fmt.Sprintf("Failed to get file info: %v", err))
		}
		lastModified := fileInfo.ModTime()

		// If the file hasn't been written in 10m, restart.
		timeDiff := time.Since(lastModified)
		printAndAddToLog(fmt.Sprintf("Last heartbeat received %s ago", timeDiff))
		if timeDiff > noResponseThresholdMinutes * time.Minute {
			printAndAddToLog(fmt.Sprintf("Restarting due to no heartbeat received in %d minutes.", noResponseThresholdMinutes))
			restart()
		}

		// Read the contents of the text file
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			printAndAddToLog(fmt.Sprintf("Failed to read file: %v", err))
		}

		// Read only the first line of the text file, cast to integer.
		line := strings.TrimSpace(string(content))
		heartbeat, err := strconv.Atoi(line)
		printAndAddToLog(fmt.Sprintf("Reported heartbeat: %d", heartbeat))
		if err != nil {
			printAndAddToLog(fmt.Sprintf("Failed to read heartbeat: %s\n", line))
			time.Sleep(60 * time.Second)
			continue
		}

		// If heartbeat is greater than 1000s, restart machine.
		if heartbeat > pulseDeadThresholdSeconds {
			printAndAddToLog(fmt.Sprintf("Restarting due to heartbeat > %d seconds.", pulseDeadThresholdSeconds))
			restart()
		}

		log.Print("Monitoring hearbeat...")
		time.Sleep(5*time.Minute)
	}
}
