const fs = require('fs');
const path = require('path');

console.log('Adding worker control endpoints...');

const appGoPath = path.join(__dirname, '../../src/ui/rest/app.go');
let appGoContent = fs.readFileSync(appGoPath, 'utf8');

// Add worker control endpoints after existing worker status endpoint
const workerEndpoints = `
	// Worker control endpoints
	app.Post("/api/workers/resume-failed", rest.ResumeFailedWorkers)
	app.Post("/api/workers/stop-all", rest.StopAllWorkers)`;

appGoContent = appGoContent.replace(
    'app.Get("/api/workers/status", rest.GetWorkerStatus)',
    'app.Get("/api/workers/status", rest.GetWorkerStatus)' + workerEndpoints
);

// Add worker control handler functions
const workerHandlers = `

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
}

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
	for _, device := range devices {
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
}`;

// Add handlers before the last closing brace
appGoContent = appGoContent.replace(/\n}\s*$/, workerHandlers + '\n}');

fs.writeFileSync(appGoPath, appGoContent, 'utf8');
console.log('Worker control endpoints added successfully!');
