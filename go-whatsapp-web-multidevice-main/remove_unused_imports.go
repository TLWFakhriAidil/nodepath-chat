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
		panic(err)
	}

	text := string(content)
	
	// Remove unused imports
	text = strings.Replace(text, `	"database/sql"
`, "", 1)
	text = strings.Replace(text, `	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
`, "", 1)

	// Write back
	err = ioutil.WriteFile("src/usecase/ultra_optimized_broadcast_processor.go", []byte(text), 0644)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Removed unused imports")
}
