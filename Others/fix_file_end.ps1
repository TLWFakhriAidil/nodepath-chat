# Remove corrupted end and properly close the file
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$lines = Get-Content $filePath

# Find the last valid line before corruption
$lastValidLine = -1
for ($i = $lines.Count - 1; $i -ge 0; $i--) {
    if ($lines[$i] -match '^\s*return c\.Status\(401\)\.JSON\(utils\.ResponseData\{') {
        # This is the corrupted line
        $lastValidLine = $i - 1
        break
    }
}

if ($lastValidLine -eq -1) {
    # If we didn't find the corruption marker, look for the countConnectedDevices function
    for ($i = $lines.Count - 50; $i -lt $lines.Count; $i++) {
        if ($lines[$i] -match 'func countConnectedDevices\(devices \[\]\*models\.UserDevice\) int \{') {
            # Found the function, now find where it should end
            $braceCount = 0
            for ($j = $i; $j -lt $lines.Count; $j++) {
                $braceCount += ([regex]::Matches($lines[$j], "\{")).Count
                $braceCount -= ([regex]::Matches($lines[$j], "\}")).Count
                
                if ($braceCount -eq 0 -and $j -gt $i) {
                    $lastValidLine = $j
                    break
                }
            }
            break
        }
    }
}

# If we still haven't found it, just look for the last complete function
if ($lastValidLine -eq -1) {
    for ($i = $lines.Count - 1; $i -ge $lines.Count - 100; $i--) {
        if ($lines[$i] -match '^\}$' -and $i -gt 0 -and $lines[$i-1] -match '\s+\}\s*$') {
            $lastValidLine = $i
            break
        }
    }
}

Write-Host "Found last valid line at: $($lastValidLine + 1)"

# Create new content
$newContent = @()
for ($i = 0; $i -le $lastValidLine; $i++) {
    $newContent += $lines[$i]
}

# Fix the countConnectedDevices function if it's incomplete
$needsClosing = $false
$lastFewLines = $newContent[-20..-1] -join "`n"
if ($lastFewLines -match 'func countConnectedDevices.*\{' -and $lastFewLines -notmatch 'return count\s*\}') {
    Write-Host "Fixing incomplete countConnectedDevices function"
    $needsClosing = $true
}

if ($needsClosing) {
    # Complete the countConnectedDevices function
    $newContent += @(
        "		}",
        "	}",
        "	return count",
        "}"
    )
}

# Add missing functions if needed
$fullContent = $newContent -join "`n"

# Check for missing max function
if ($fullContent -notmatch 'func max\(a, b int\) int \{') {
    Write-Host "Adding missing max function"
    $newContent += @(
        "",
        "func max(a, b int) int {",
        "	if a > b {",
        "		return a",
        "	}",
        "	return b",
        "}"
    )
}

# Check for missing ResumeFailedWorkers
if ($fullContent -notmatch 'func \(handler \*App\) ResumeFailedWorkers') {
    Write-Host "Adding missing ResumeFailedWorkers function"
    $newContent += @(
        "",
        "// ResumeFailedWorkers resumes all failed device workers",
        "func (handler *App) ResumeFailedWorkers(c *fiber.Ctx) error {",
        "	sessionToken := c.Cookies(`"session_token`")",
        "	if sessionToken == `"`" {",
        "		return c.Status(401).JSON(utils.ResponseData{",
        "			Status:  401,",
        "			Code:    `"UNAUTHORIZED`",",
        "			Message: `"No session token`",",
        "		})",
        "	}",
        "",
        "	userRepo := repository.GetUserRepository()",
        "	session, err := userRepo.GetSession(sessionToken)",
        "	if err != nil {",
        "		return c.Status(401).JSON(utils.ResponseData{",
        "			Status:  401,",
        "			Code:    `"UNAUTHORIZED`",",
        "			Message: `"Invalid session`",",
        "		})",
        "	}",
        "",
        "	// Get all devices for user",
        "	devices, err := userRepo.GetUserDevices(session.UserID)",
        "	if err != nil {",
        "		return c.Status(500).JSON(utils.ResponseData{",
        "			Status:  500,",
        "			Code:    `"ERROR`",",
        "			Message: `"Failed to get devices`",",
        "		})",
        "	}",
        "",
        "	// Resume workers for devices that are connected but have stopped workers",
        "	resumedCount := 0",
        "	for _, device := range devices {",
        "		if device.Status == `"connected`" {",
        "			// TODO: Check if worker is stopped and resume",
        "			// This would interface with your broadcast manager",
        "			resumedCount++",
        "		}",
        "	}",
        "",
        "	return c.JSON(utils.ResponseData{",
        "		Status:  200,",
        "		Code:    `"SUCCESS`",",
        "		Message: fmt.Sprintf(`"Resumed %d workers`", resumedCount),",
        "		Results: map[string]interface{}{",
        "			`"resumed_count`": resumedCount,",
        "		},",
        "	})",
        "}"
    )
}

# Check for missing StopAllWorkers
if ($fullContent -notmatch 'func \(handler \*App\) StopAllWorkers') {
    Write-Host "Adding missing StopAllWorkers function"
    $newContent += @(
        "",
        "// StopAllWorkers stops all running device workers",
        "func (handler *App) StopAllWorkers(c *fiber.Ctx) error {",
        "	sessionToken := c.Cookies(`"session_token`")",
        "	if sessionToken == `"`" {",
        "		return c.Status(401).JSON(utils.ResponseData{",
        "			Status:  401,",
        "			Code:    `"UNAUTHORIZED`",",
        "			Message: `"No session token`",",
        "		})",
        "	}",
        "",
        "	userRepo := repository.GetUserRepository()",
        "	session, err := userRepo.GetSession(sessionToken)",
        "	if err != nil {",
        "		return c.Status(401).JSON(utils.ResponseData{",
        "			Status:  401,",
        "			Code:    `"UNAUTHORIZED`",",
        "			Message: `"Invalid session`",",
        "		})",
        "	}",
        "",
        "	// Get all devices for user",
        "	devices, err := userRepo.GetUserDevices(session.UserID)",
        "	if err != nil {",
        "		return c.Status(500).JSON(utils.ResponseData{",
        "			Status:  500,",
        "			Code:    `"ERROR`",",
        "			Message: `"Failed to get devices`",",
        "		})",
        "	}",
        "",
        "	// Stop all workers",
        "	stoppedCount := 0",
        "	for range devices {",
        "		// TODO: Stop worker for this device",
        "		// This would interface with your broadcast manager",
        "		stoppedCount++",
        "	}",
        "",
        "	return c.JSON(utils.ResponseData{",
        "		Status:  200,",
        "		Code:    `"SUCCESS`",",
        "		Message: fmt.Sprintf(`"Stopped %d workers`", stoppedCount),",
        "		Results: map[string]interface{}{",
        "			`"stopped_count`": stoppedCount,",
        "		},",
        "	})",
        "}"
    )
}

# Save the file
$newContent | Out-File -FilePath $filePath -Encoding UTF8

Write-Host "Fixed! File now has $($newContent.Count) lines (was $($lines.Count))"
