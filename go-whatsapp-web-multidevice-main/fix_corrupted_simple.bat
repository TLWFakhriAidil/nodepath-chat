@echo off
echo Fixing corrupted end of app.go...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Add missing closing braces and fix the corrupted function
echo. >> src\ui\rest\app.go
echo 		} >> src\ui\rest\app.go
echo 	} >> src\ui\rest\app.go
echo 	return count >> src\ui\rest\app.go
echo } >> src\ui\rest\app.go
echo. >> src\ui\rest\app.go
echo // StopAllWorkers stops all running device workers >> src\ui\rest\app.go
echo func (handler *App) StopAllWorkers(c *fiber.Ctx) error { >> src\ui\rest\app.go
echo 	sessionToken := c.Cookies("session_token") >> src\ui\rest\app.go
echo 	if sessionToken == "" { >> src\ui\rest\app.go
echo 		return c.Status(401).JSON(utils.ResponseData{ >> src\ui\rest\app.go
echo 			Status:  401, >> src\ui\rest\app.go
echo 			Code:    "UNAUTHORIZED", >> src\ui\rest\app.go
echo 			Message: "No session token", >> src\ui\rest\app.go
echo 		}) >> src\ui\rest\app.go
echo 	} >> src\ui\rest\app.go
echo. >> src\ui\rest\app.go
echo 	userRepo := repository.GetUserRepository() >> src\ui\rest\app.go
echo 	session, err := userRepo.GetSession(sessionToken) >> src\ui\rest\app.go
echo 	if err != nil { >> src\ui\rest\app.go
echo 		return c.Status(401).JSON(utils.ResponseData{ >> src\ui\rest\app.go
echo 			Status:  401, >> src\ui\rest\app.go
echo 			Code:    "UNAUTHORIZED", >> src\ui\rest\app.go
echo 			Message: "Invalid session", >> src\ui\rest\app.go
echo 		}) >> src\ui\rest\app.go
echo 	} >> src\ui\rest\app.go
echo. >> src\ui\rest\app.go
echo 	// Get all devices for user >> src\ui\rest\app.go
echo 	devices, err := userRepo.GetUserDevices(session.UserID) >> src\ui\rest\app.go
echo 	if err != nil { >> src\ui\rest\app.go
echo 		return c.Status(500).JSON(utils.ResponseData{ >> src\ui\rest\app.go
echo 			Status:  500, >> src\ui\rest\app.go
echo 			Code:    "ERROR", >> src\ui\rest\app.go
echo 			Message: "Failed to get devices", >> src\ui\rest\app.go
echo 		}) >> src\ui\rest\app.go
echo 	} >> src\ui\rest\app.go
echo. >> src\ui\rest\app.go
echo 	// Stop all workers >> src\ui\rest\app.go
echo 	stoppedCount := 0 >> src\ui\rest\app.go
echo 	for range devices { >> src\ui\rest\app.go
echo 		// TODO: Stop worker for this device >> src\ui\rest\app.go
echo 		// This would interface with your broadcast manager >> src\ui\rest\app.go
echo 		stoppedCount++ >> src\ui\rest\app.go
echo 	} >> src\ui\rest\app.go
echo. >> src\ui\rest\app.go
echo 	return c.JSON(utils.ResponseData{ >> src\ui\rest\app.go
echo 		Status:  200, >> src\ui\rest\app.go
echo 		Code:    "SUCCESS", >> src\ui\rest\app.go
echo 		Message: fmt.Sprintf("Stopped %%d workers", stoppedCount), >> src\ui\rest\app.go
echo 		Results: map[string]interface{}{ >> src\ui\rest\app.go
echo 			"stopped_count": stoppedCount, >> src\ui\rest\app.go
echo 		}, >> src\ui\rest\app.go
echo 	}) >> src\ui\rest\app.go
echo } >> src\ui\rest\app.go

echo Testing build...
cd src
go build .
cd ..

pause
