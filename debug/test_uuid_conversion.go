package main

import (
	"fmt"
	"hash/fnv"
)

// convertUUIDToInt converts a UUID string to an integer using FNV-32a hash
func convertUUIDToInt(uuid string) int {
	h := fnv.New32a()
	h.Write([]byte(uuid))
	return int(h.Sum32())
}

func main() {
	// Test with existing user UUIDs from the database
	userUUIDs := []string{
		"114fdace4199f52e8ac6bc49955d32b2",
		"18c2c749-6935-47cf-a172-46112bd7c1a9",
		"1f2b41476cf4a186d1b40245b542cbd2",
		"5cbabc10f3c98b8fd2dfdf1074f3091c",
		"98480f1c-413e-4aea-9a66-dc5b2181bc59",
	}

	fmt.Println("UUID to Integer conversions:")
	for _, uuid := range userUUIDs {
		convertedID := convertUUIDToInt(uuid)
		fmt.Printf("UUID: %s -> Integer: %d\n", uuid, convertedID)
	}

	// Check what the current device is linked to
	fmt.Printf("\nDevice FakhriAidilTLW-001 is linked to user ID: 1727068808\n")
	fmt.Println("This doesn't match any of the converted UUIDs above.")
}