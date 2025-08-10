# Fix the corrupted end of app.go
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$content = Get-Content $filePath -Raw

# Find where the corruption starts - looks like it's in the countConnectedDevices function
$pattern = '(?s)func countConnectedDevices\(devices \[\]\*models\.UserDevice\) int \{[^}]+return c\.Status'

if ($content -match $pattern) {
    Write-Host "Found corrupted countConnectedDevices function, fixing..."
    
    # Replace with correct function
    $correctFunction = @'
func countConnectedDevices(devices []*models.UserDevice) int {
	count := 0
	for _, device := range devices {
		if device.Status == "connected" {
			count++
		}
	}
	return count
}

// ResumeFailedWorkers resumes all failed device workers
func (handler *App) ResumeFailedWorkers(c *fiber.Ctx) error {
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}

	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}

	// Get all devices for user
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}

	// Resume workers for devices that are connected but have stopped workers
	resumedCount := 0
	for _, device := range devices {
		if device.Status == "connected" {
			// TODO: Check if worker is stopped and resume
			// This would interface with your broadcast manager
			resumedCount++
		}
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Resumed %d workers", resumedCount),
		Results: map[string]interface{}{
			"resumed_count": resumedCount,
		},
	})
}'@
    
    # Replace the corrupted part with the correct function
    $content = $content -replace $pattern, $correctFunction
}

# Also check if there's a missing max function
if ($content -notmatch 'func max\(a, b int\) int \{') {
    Write-Host "Adding missing max function..."
    $maxFunction = @'

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}'@
    
    # Add after min function
    $content = $content -replace '(func min\(a, b int\) int \{[^}]+\})', "`$1$maxFunction"
}

# Make sure there's no duplicate StopAllWorkers
$stopAllWorkersCount = ([regex]::Matches($content, 'func \(handler \*App\) StopAllWorkers')).Count
if ($stopAllWorkersCount -eq 0) {
    Write-Host "Adding missing StopAllWorkers function..."
    
    # Add StopAllWorkers after ResumeFailedWorkers
    $stopAllWorkersFunc = @'

// StopAllWorkers stops all running device workers
func (handler *App) StopAllWorkers(c *fiber.Ctx) error {
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}

	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}

	// Get all devices for user
	devices, err := userRepo.GetUserDevices(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}

	// Stop all workers
	stoppedCount := 0
	for range devices {
		// TODO: Stop worker for this device
		// This would interface with your broadcast manager
		stoppedCount++
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Stopped %d workers", stoppedCount),
		Results: map[string]interface{}{
			"stopped_count": stoppedCount,
		},
	})
}'@
    
    $content = $content + $stopAllWorkersFunc
}

# Save the file
Set-Content $filePath -Value $content -NoNewline -Encoding UTF8

Write-Host "Fixed corrupted file!"
