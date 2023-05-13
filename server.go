package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	filePath := "heartbeat.txt"

	for {
		// Check if the text file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("File does not exist: %s\n", filePath)
			time.Sleep(5 * time.Second)
			continue
		}

		// Read the contents of the text file
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		// Convert the file contents into a string to integer mapping
		data := make(map[string]int)
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if intValue, err := strconv.Atoi(value); err == nil {
					data[key] = intValue
				} else {
					log.Printf("Failed to convert value to integer: %s\n", value)
				}
			}
		}

		key1Value, _ := data["key1"]
		key2Value, _ := data["key2"]
		if key1Value > 5000 ||  key2Value > 5000{
			fileContents := ""
			for key := range data {
				fileContents += key + " = 0\n"
			}

			// Overwrite the original file with the updated contents
			err := ioutil.WriteFile(filePath, []byte(fileContents), 0644)
			if err != nil {
				log.Fatalf("Failed to write file: %v", err)
			}

			// Restart the computer
			if err := exec.Command("sudo", "reboot").Run(); err != nil {
				log.Fatalf("Failed to reboot: %v", err)
			}
		}else{
			log.Print("Heartbeat checking...")
		}

		time.Sleep(60 * time.Second)
	}
}