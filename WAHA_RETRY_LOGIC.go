// WAHA ERROR HANDLING IMPROVEMENT
// This adds retry logic without unnecessary session restarts
// Add this to the sendWahaTextMessage function in waha_support.go

// Replace the current error handling in sendWahaTextMessage with:

	// Send request with retry mechanism
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		
		if err != nil {
			if attempt < maxRetries {
				logrus.WithFields(logrus.Fields{
					"attempt": attempt,
					"error":   err.Error(),
				}).Warn("⚠️ WAHA: Request failed, retrying...")
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
				continue
			}
			logrus.WithError(err).Error("❌ WAHA: Failed after retries")
			return
		}
		
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		
		// Success
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			logrus.Info("✅ WAHA: Message sent successfully")
			return
		}
		
		// Server error - retry with longer delay
		if resp.StatusCode == http.StatusInternalServerError && attempt < maxRetries {
			logrus.WithFields(logrus.Fields{
				"attempt": attempt,
				"error":   string(body),
			}).Warn("⚠️ WAHA: Server error, retrying...")
			time.Sleep(time.Duration(attempt) * 3 * time.Second)
			
			// Recreate request for retry
			req, _ = http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Api-Key", apiKey)
			continue
		}
		
		// Failed
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
		}).Error("❌ WAHA: Failed to send message")
		return
	}