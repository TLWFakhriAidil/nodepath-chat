package rest

import (
	"database/sql"
	"time"
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// InitTeamRoutes initializes team member routes
func InitTeamRoutes(app *fiber.App, db *sql.DB) {
	log.Println("Initializing team routes...")
	
	// Team login page
	app.Get("/team/login", func(c *fiber.Ctx) error {
		return c.SendFile("./views/team_login.html")
	})

	// Team dashboard page
	app.Get("/team/dashboard", func(c *fiber.Ctx) error {
		return c.SendFile("./views/team_dashboard.html")
	})
	
	// Debug endpoint to test if routes are loaded
	app.Get("/api/team/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "Team routes are loaded!"})
	})

	// Team login API
	app.Post("/api/team/login", func(c *fiber.Ctx) error {
		log.Println("Team login endpoint called")
		
		var loginReq struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&loginReq); err != nil {
			log.Printf("Failed to parse team login request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
				"debug": err.Error(),
			})
		}
		
		log.Printf("Team login attempt for username: %s", loginReq.Username)
		
		// Log the request for debugging
		if loginReq.Username == "" || loginReq.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
				"debug": "Empty username or password",
			})
		}

		// Check credentials - also check is_active
		var memberID int
		var username string
		var isActive bool
		err := db.QueryRow(`
			SELECT id, username, is_active FROM team_members 
			WHERE username = ? AND password = ?
		`, loginReq.Username, loginReq.Password).Scan(&memberID, &username, &isActive)

		if err != nil {
			// Log for debugging
			if err == sql.ErrNoRows {
				// Check if user exists
				var count int
				db.QueryRow("SELECT COUNT(*) from team_members WHERE username = ?", loginReq.Username).Scan(&count)
				if count > 0 {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "Invalid credentials - password mismatch",
						"debug": "User exists but password doesn't match",
					})
				}
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid credentials - user not found",
					"debug": "No user with username: " + loginReq.Username,
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials", 
				"debug": err.Error(),
			})
		}

		// Check if active
		if !isActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Account is inactive",
				"debug": "User exists but is_active = false",
			})
		}

		// Create session
		sessionID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		_, err = db.Exec(`
			INSERT INTO team_sessions (team_member_id, session_id, expires_at, created_at)
			VALUES (?, ?, ?, NOW())
		`, memberID, sessionID, expiresAt)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
		}

		// Set cookie
		c.Cookie(&fiber.Cookie{
			Name:     "team_session",
			Value:    sessionID,
			Expires:  expiresAt,
			HTTPOnly: true,
			Path:     "/",
		})

		return c.JSON(fiber.Map{
			"status": "success",
			"token": sessionID,
			"user": fiber.Map{
				"username": username,
				"device_name": username,
			},
		})
	})

	// Team API group with auth middleware
	teamAPI := app.Group("/api/team", TeamAuthMiddleware(db))

	// Team member info
	teamAPI.Get("/member-info", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)

		return c.JSON(fiber.Map{
			"code": "SUCCESS",
			"results": fiber.Map{
				"username":    username,
				"device_name": username, // Username matches device_name
			},
		})
	})

	// Team logout
	teamAPI.Post("/logout", func(c *fiber.Ctx) error {
		sessionID := c.Cookies("team_session")
		if sessionID != "" {
			db.Exec("DELETE FROM team_sessions WHERE session_id = ?", sessionID)
		}
		
		c.Cookie(&fiber.Cookie{
			Name:     "team_session",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HTTPOnly: true,
			Path:     "/",
		})
		
		return c.JSON(fiber.Map{"success": true})
	})

	// Devices - only show device where device_name matches username
	teamAPI.Get("/devices", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)
		
		rows, err := db.Query(`
			SELECT id, device_name, phone, status, jid, last_seen
			FROM user_devices
			WHERE device_name = ?
		`, username)
		
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch devices"})
		}
		defer rows.Close()

		var devices []fiber.Map
		for rows.Next() {
			var device struct {
				ID       string
				Name     string
				Phone    sql.NullString
				Status   string
				JID      sql.NullString
				LastSeen sql.NullTime
			}
			rows.Scan(&device.ID, &device.Name, &device.Phone, &device.Status, &device.JID, &device.LastSeen)
			
			devices = append(devices, fiber.Map{
				"id":       device.ID,
				"name":     device.Name,
				"phone":    device.Phone.String,
				"status":   device.Status,
				"jid":      device.JID.String,
				"lastSeen": device.LastSeen.Time,
			})
		}

		return c.JSON(fiber.Map{
			"code":    "SUCCESS",
			"results": devices,
		})
	})

	// Campaign analytics - filtered by device
	teamAPI.Get("/campaigns/analytics", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)
		startDate := c.Query("start")
		endDate := c.Query("end")
		niche := c.Query("niche")

		// Base query filtered by device
		query := `
			SELECT COUNT(DISTINCT c.id) AS total_campaigns,
				COUNT(DISTINCT bm.id) AS total_contacts_should_send,
				COUNT(DISTINCT case WHEN bm.status = 'sent' THEN bm.id END) AS contacts_done_send,
				COUNT(DISTINCT case WHEN bm.status = 'failed' THEN bm.id END) AS contacts_failed_send,
				COUNT(DISTINCT case WHEN bm.status = 'pending' THEN bm.id END) AS contacts_remaining_send
			FROM campaigns c
			LEFT JOIN broadcast_messages bm ON c.id = bm.campaign_id
			WHERE c.device_id IN (SELECT id FROM user_devices WHERE device_name = ?)
		`
		args := []interface{}{username}
		paramCount := 1

		if startDate != "" && endDate != "" {
			paramCount++
			query += " AND DATE(c.date) BETWEEN $" + string(rune('0'+paramCount))
			paramCount++
			query += " AND $" + string(rune('0'+paramCount))
			args = append(args, startDate, endDate)
		}

		if niche != "" && niche != "all" {
			paramCount++
			query += " AND c.niche = $" + string(rune('0'+paramCount))
			args = append(args, niche)
		}

		var analytics struct {
			TotalCampaigns          int
			TotalContactsShouldSend int
			ContactsDoneSend        int
			ContactsFailedSend      int
			ContactsRemainingSend   int
		}

		err := db.QueryRow(query, args...).Scan(
			&analytics.TotalCampaigns,
			&analytics.TotalContactsShouldSend,
			&analytics.ContactsDoneSend,
			&analytics.ContactsFailedSend,
			&analytics.ContactsRemainingSend,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch analytics"})
		}

		// Get chart data
		chartQuery := `
			SELECT DATE(bm.sent_at) AS date,
				COUNT(CASE WHEN bm.status = 'sent' THEN 1 END) AS sent,
				COUNT(CASE WHEN bm.status = 'failed' THEN 1 END) AS failed
			FROM broadcast_messages bm
			JOIN campaigns c ON bm.campaign_id = c.id
			WHERE c.device_id IN (SELECT id FROM user_devices WHERE device_name = ?)
		`
		chartArgs := []interface{}{username}
		chartParamCount := 1

		if startDate != "" && endDate != "" {
			chartParamCount++
			chartQuery += " AND DATE(bm.sent_at) BETWEEN $" + string(rune('0'+chartParamCount))
			chartParamCount++
			chartQuery += " AND $" + string(rune('0'+chartParamCount))
			chartArgs = append(chartArgs, startDate, endDate)
		}

		chartQuery += " GROUP BY DATE(bm.sent_at) ORDER BY date"

		rows, _ := db.Query(chartQuery, chartArgs...)
		defer rows.Close()

		var labels []string
		var sentData []int
		var failedData []int

		for rows.Next() {
			var date string
			var sent, failed int
			rows.Scan(&date, &sent, &failed)
			labels = append(labels, date)
			sentData = append(sentData, sent)
			failedData = append(failedData, failed)
		}

		return c.JSON(fiber.Map{
			"totalCampaigns":          analytics.TotalCampaigns,
			"totalContactsShouldSend": analytics.TotalContactsShouldSend,
			"contactsDoneSend":        analytics.ContactsDoneSend,
			"contactsFailedSend":      analytics.ContactsFailedSend,
			"contactsRemainingSend":   analytics.ContactsRemainingSend,
			"chartData": fiber.Map{
				"labels": labels,
				"sent":   sentData,
				"failed": failedData,
			},
		})
	})

	// Sequence analytics - filtered by device
	teamAPI.Get("/sequences/analytics", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)
		startDate := c.Query("start")
		endDate := c.Query("end")
		niche := c.Query("niche")

		// Base query filtered by device
		query := `
			SELECT COUNT(DISTINCT s.id) AS total_sequences,
				COUNT(DISTINCT ss.id) AS total_flows,
				COUNT(DISTINCT sc.id) AS total_contacts_should_send,
				COUNT(DISTINCT case WHEN sc.status = 'completed' THEN sc.id END) AS contacts_done_send,
				COUNT(DISTINCT case WHEN sc.status = 'failed' THEN sc.id END) AS contacts_failed_send,
				COUNT(DISTINCT case WHEN sc.status = 'pending' THEN sc.id END) AS contacts_remaining_send
			FROM sequences s
			LEFT JOIN sequence_steps ss ON s.id = ss.sequence_id
			LEFT JOIN sequence_contacts sc ON s.id = sc.sequence_id
			WHERE sc.device_id IN (SELECT id FROM user_devices WHERE device_name = ?)
		`
		args := []interface{}{username}
		paramCount := 1

		if startDate != "" && endDate != "" {
			paramCount++
			query += " AND DATE(sc.completed_at) BETWEEN $" + string(rune('0'+paramCount))
			paramCount++
			query += " AND $" + string(rune('0'+paramCount))
			args = append(args, startDate, endDate)
		}

		if niche != "" && niche != "all" {
			paramCount++
			query += " AND s.niche = $" + string(rune('0'+paramCount))
			args = append(args, niche)
		}

		var analytics struct {
			TotalSequences          int
			TotalFlows             int
			TotalContactsShouldSend int
			ContactsDoneSend        int
			ContactsFailedSend      int
			ContactsRemainingSend   int
		}

		err := db.QueryRow(query, args...).Scan(
			&analytics.TotalSequences,
			&analytics.TotalFlows,
			&analytics.TotalContactsShouldSend,
			&analytics.ContactsDoneSend,
			&analytics.ContactsFailedSend,
			&analytics.ContactsRemainingSend,
		)

		if err != nil {
			// Return empty data if error
			analytics = struct {
				TotalSequences          int
				TotalFlows             int
				TotalContactsShouldSend int
				ContactsDoneSend        int
				ContactsFailedSend      int
				ContactsRemainingSend   int
			}{}
		}

		// Get chart data
		var labels []string
		var completedData []int
		var failedData []int
		var pendingData []int

		return c.JSON(fiber.Map{
			"totalSequences":          analytics.TotalSequences,
			"totalFlows":             analytics.TotalFlows,
			"totalContactsShouldSend": analytics.TotalContactsShouldSend,
			"contactsDoneSend":        analytics.ContactsDoneSend,
			"contactsFailedSend":      analytics.ContactsFailedSend,
			"contactsRemainingSend":   analytics.ContactsRemainingSend,
			"chartData": fiber.Map{
				"labels":    labels,
				"completed": completedData,
				"failed":    failedData,
				"pending":   pendingData,
			},
		})
	})

	// Campaign summary - filtered by device
	teamAPI.Get("/campaigns/summary", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)
		
		rows, err := db.Query(`
			SELECT c.id, c.title, c.date, c.niche, c.status
			FROM campaigns c
			JOIN user_devices ud ON c.device_id = ud.id
			WHERE ud.device_name = ?
			ORDER BY c.date DESC
			LIMIT 50
		`, username)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch campaigns"})
		}
		defer rows.Close()

		var campaigns []fiber.Map
		for rows.Next() {
			var campaign struct {
				ID     string
				Title  string
				Date   string
				Niche  string
				Status string
			}
			rows.Scan(&campaign.ID, &campaign.Title, &campaign.Date, &campaign.Niche, &campaign.Status)
			campaigns = append(campaigns, fiber.Map{
				"id":     campaign.ID,
				"title":  campaign.Title,
				"date":   campaign.Date,
				"niche":  campaign.Niche,
				"status": campaign.Status,
			})
		}

		return c.JSON(campaigns)
	})

	// Niches - filtered by device campaigns
	teamAPI.Get("/niches", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)
		
		rows, err := db.Query(`
			SELECT DISTINCT c.niche
			FROM campaigns c
			JOIN user_devices ud ON c.device_id = ud.id
			WHERE ud.device_name = ? AND c.niche IS NOT NULL AND c.niche != ''
			ORDER BY c.niche
		`, username)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch niches"})
		}
		defer rows.Close()

		var niches []string
		for rows.Next() {
			var niche string
			rows.Scan(&niche)
			niches = append(niches, niche)
		}

		return c.JSON(niches)
	})

	// Sequences list - filtered by device
	teamAPI.Get("/sequences", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)
		
		rows, err := db.Query(`
			SELECT DISTINCT s.id, s.name, s.description, s.niche, s.trigger_name, s.is_active
			FROM sequences s
			WHERE EXISTS (
				SELECT 1 FROM sequence_contacts sc
				WHERE sc.sequence_id = s.id
				AND sc.device_id IN (SELECT id FROM user_devices WHERE device_name = ?)
			)
			ORDER BY s.created_at DESC
		`, username)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch sequences"})
		}
		defer rows.Close()

		var sequences []fiber.Map
		for rows.Next() {
			var sequence struct {
				ID          string
				Name        string
				Description string
				Niche       sql.NullString
				TriggerName sql.NullString
				IsActive    bool
			}
			rows.Scan(&sequence.ID, &sequence.Name, &sequence.Description, &sequence.Niche, &sequence.TriggerName, &sequence.IsActive)
			sequences = append(sequences, fiber.Map{
				"id":           sequence.ID,
				"name":         sequence.Name,
				"description":  sequence.Description,
				"niche":        sequence.Niche.String,
				"trigger_name": sequence.TriggerName.String,
				"is_active":    sequence.IsActive,
			})
		}

		return c.JSON(fiber.Map{
			"code":    "SUCCESS",
			"results": sequences,
		})
	})

	// Sequence summary - filtered by device
	teamAPI.Get("/sequences/summary", func(c *fiber.Ctx) error {
		username := c.Locals("team_username").(string)
		
		rows, err := db.Query(`
			SELECT s.id, 
				s.name, 
				s.niche,
				COUNT(DISTINCT sc.id) AS total_contacts,
				COUNT(DISTINCT case WHEN sc.status = 'completed' THEN sc.id END) AS completed,
				COUNT(DISTINCT case WHEN sc.status = 'failed' THEN sc.id END) AS failed,
				COUNT(DISTINCT case WHEN sc.status = 'pending' THEN sc.id END) AS pending
			FROM sequences s
			LEFT JOIN sequence_contacts sc ON s.id = sc.sequence_id
			WHERE sc.device_id IN (SELECT id FROM user_devices WHERE device_name = ?)
			GROUP BY s.id, s.name, s.niche
			ORDER BY s.created_at DESC
		`, username)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch sequence summary"})
		}
		defer rows.Close()

		var summaries []fiber.Map
		for rows.Next() {
			var summary struct {
				ID            string
				Name          string
				Niche         sql.NullString
				TotalContacts int
				Completed     int
				Failed        int
				Pending       int
			}
			rows.Scan(&summary.ID, &summary.Name, &summary.Niche, &summary.TotalContacts, 
				&summary.Completed, &summary.Failed, &summary.Pending)
			
			summaries = append(summaries, fiber.Map{
				"id":             summary.ID,
				"name":           summary.Name,
				"niche":          summary.Niche.String,
				"total_contacts": summary.TotalContacts,
				"completed":      summary.Completed,
				"failed":         summary.Failed,
				"pending":        summary.Pending,
			})
		}

		return c.JSON(summaries)
	})
}

// TeamAuthMiddleware checks team member authentication
func TeamAuthMiddleware(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies("team_session")
		if sessionID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No team session found"})
		}

		var username string
		err := db.QueryRow(`
			SELECT tm.username
			FROM team_sessions ts
			JOIN team_members tm ON ts.team_member_id = tm.id
			WHERE ts.session_id = ? AND ts.expires_at > NOW()
		`, sessionID).Scan(&username)

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired session"})
		}

		c.Locals("team_username", username)
		c.Locals("team_device_name", username) // Username matches device_name
		
		return c.Next()
	}
}
