package main

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

func main() {
	// Read the file
	content, err := ioutil.ReadFile("internal/repository/ai_whatsapp_repository.go")
	if err != nil {
		log.Fatal(err)
	}
	
	text := string(content)
	
	// Fix the FlowReference block
	text = strings.Replace(text, 
		`if ai.FlowReference.Valid {
		flowReferenceValue = ai.FlowReference.String
	} else {
		flowReferenceValue = nil
	}`,
		`// FlowReference removed from database
	flowReferenceValue = nil`,
		1)
	
	// Fix the ExecutionID block
	text = strings.Replace(text,
		`if ai.ExecutionID.Valid {
		executionIDValue = ai.ExecutionID.String
	} else {
		executionIDValue = nil
	}`,
		`// ExecutionID removed from database
	executionIDValue = nil`,
		1)
	
	// Fix the ExecutionStatus block
	text = strings.Replace(text,
		`if ai.ExecutionStatus.Valid {
		executionStatusValue = ai.ExecutionStatus.String
	} else {
		executionStatusValue = nil
	}`,
		`// ExecutionStatus removed from database
	executionStatusValue = nil`,
		1)
	
	// Fix the Scan calls
	text = regexp.MustCompile(`&ai\.ExecutionStatus, &ai\.ExecutionID,`).ReplaceAllString(text, 
		`// &ai.ExecutionStatus, &ai.ExecutionID, // Removed fields`)
	
	// Write back
	err = ioutil.WriteFile("internal/repository/ai_whatsapp_repository.go", []byte(text), 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Fixed ai_whatsapp_repository.go")
}