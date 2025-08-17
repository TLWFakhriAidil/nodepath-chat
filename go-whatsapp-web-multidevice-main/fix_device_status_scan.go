package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// Read the file
	content, err := ioutil.ReadFile("src/usecase/ultra_optimized_broadcast_processor.go")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fileContent := string(content)

	// Fix 1: Change deviceStatus from string to sql.NullString
	fileContent = strings.Replace(fileContent, 
		"var deviceStatus string",
		"var deviceStatus sql.NullString",
		1)

	// Fix 2: Update the usage of deviceStatus in the code
	// Find and replace where deviceStatus is used
	fileContent = strings.Replace(fileContent,
		"DeviceStatus: deviceStatus,",
		"DeviceStatus: deviceStatus.String,",
		-1)

	// Fix 3: Add COALESCE to the SQL query to handle NULL values
	fileContent = strings.Replace(fileContent,
		"d.status AS device_status,",
		"COALESCE(d.status, 'unknown') AS device_status,",
		1)

	// Write the fixed content back
	err = ioutil.WriteFile("src/usecase/ultra_optimized_broadcast_processor.go", []byte(fileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Println("Fixed device_status NULL handling successfully!")
}
