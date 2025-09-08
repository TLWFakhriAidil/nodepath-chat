package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

/**
 * Test the device status API endpoint to verify Analytics fix
 */
func main() {
	fmt.Println("=== TESTING DEVICE STATUS API ===")

	// Test login first
	accessToken, err := testLogin()
	if err != nil {
		log.Printf("❌ Login failed: %v", err)
		return
	}
	fmt.Printf("✅ Login successful, token: %s...\n", accessToken[:20])

	// Test device status endpoint
	err = testDeviceStatus(accessToken)
	if err != nil {
		log.Printf("❌ Device status test failed: %v", err)
		return
	}

	fmt.Println("\n✅ All tests passed! Analytics should now work correctly.")
}

/**
 * Test login with the created test user
 */
func testLogin() (string, error) {
	fmt.Println("\n1. Testing login with test user:")

	loginData := map[string]string{
		"email":    "test-user-1727068808@nodepath.local",
		"password": "test123456",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login data: %w", err)
	}

	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read login response: %w", err)
	}

	fmt.Printf("   Login response status: %d\n", resp.StatusCode)
	fmt.Printf("   Login response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to get token
	var loginResponse map[string]interface{}
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse login response: %w", err)
	}

	data, ok := loginResponse["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid login response format")
	}

	token, ok := data["token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in login response")
	}

	return token, nil
}

/**
 * Test device status endpoint
 */
func testDeviceStatus(accessToken string) error {
	fmt.Println("\n2. Testing device status endpoint:")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/api/auth/device-status", nil)
	if err != nil {
		return fmt.Errorf("failed to create device status request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make device status request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read device status response: %w", err)
	}

	fmt.Printf("   Device status response status: %d\n", resp.StatusCode)
	fmt.Printf("   Device status response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("device status failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to check device data
	var deviceResponse map[string]interface{}
	err = json.Unmarshal(body, &deviceResponse)
	if err != nil {
		return fmt.Errorf("failed to parse device status response: %w", err)
	}

	// Check if has_devices is true
	hasDevices, ok := deviceResponse["has_devices"].(bool)
	if !ok {
		return fmt.Errorf("has_devices field not found or invalid type")
	}

	deviceCount, ok := deviceResponse["device_count"].(float64)
	if !ok {
		return fmt.Errorf("device_count field not found or invalid type")
	}

	deviceIDs, ok := deviceResponse["device_ids"].([]interface{})
	if !ok {
		return fmt.Errorf("device_ids field not found or invalid type")
	}

	fmt.Printf("   ✅ has_devices: %t\n", hasDevices)
	fmt.Printf("   ✅ device_count: %.0f\n", deviceCount)
	fmt.Printf("   ✅ device_ids: %v\n", deviceIDs)

	if !hasDevices {
		return fmt.Errorf("has_devices is false - Analytics will still show 'No devices available'")
	}

	if deviceCount == 0 {
		return fmt.Errorf("device_count is 0 - Analytics will still show 'No devices available'")
	}

	if len(deviceIDs) == 0 {
		return fmt.Errorf("device_ids is empty - Analytics will still show 'No devices available'")
	}

	fmt.Println("   ✅ Device status API is working correctly!")
	return nil
}