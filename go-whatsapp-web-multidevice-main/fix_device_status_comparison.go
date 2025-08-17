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

	// Since we already changed deviceStatus to sql.NullString, 
	// we need to fix the comparisons
	
	// Fix the comparison operations
	fileContent = strings.Replace(fileContent,
		`if devicePlatform == "" && deviceStatus != "connected" && deviceStatus != "online" {`,
		`if devicePlatform == "" && deviceStatus.String != "connected" && deviceStatus.String != "online" {`,
		1)

	// Write the fixed content back
	err = ioutil.WriteFile("src/usecase/ultra_optimized_broadcast_processor.go", []byte(fileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Println("Fixed device_status comparisons successfully!")
}
