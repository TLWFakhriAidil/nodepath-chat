// WAHA WEBSOCKET FIX
// Add this enhanced error handling to waha_support.go to handle websocket disconnections

// Add this function to handle session restart:
func (h *Handlers) handleWahaWebsocketError(sessionName string, req *http.Request) error {
	apiBase := "https://waha-plus-production-705f.up.railway.app"
	apiKey := "dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0"
	
	// First, try to restart the session
	restartURL := fmt.Sprintf("%s/api/sessions/%s/restart", apiBase, sessionName)
	restartReq, err := http.NewRequest("POST", restartURL, nil)
	if err != nil {
		return err
	}
	restartReq.Header.Set("X-Api-Key", apiKey)
	
	client := &http.Client{Timeout: 30 * time.Second}
	restartResp, err := client.Do(restartReq)
	if err != nil {
		// If restart fails, try to start the session
		startURL := fmt.Sprintf("%s/api/sessions/%s/start", apiBase, sessionName)
		startReq, _ := http.NewRequest("POST", startURL, nil)
		startReq.Header.Set("X-Api-Key", apiKey)
		
		startResp, err := client.Do(startReq)
		if err != nil {
			return fmt.Errorf("failed to restart/start session: %w", err)
		}
		defer startResp.Body.Close()
	} else {
		defer restartResp.Body.Close()
	}
	
	// Wait for session to initialize
	time.Sleep(5 * time.Second)
	
	// Retry the original request
	retryResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to retry after restart: %w", err)
	}
	defer retryResp.Body.Close()
	
	retryBody, _ := io.ReadAll(retryResp.Body)
	if retryResp.StatusCode != http.StatusOK && retryResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("retry failed with status %d: %s", retryResp.StatusCode, string(retryBody))
	}
	
	logrus.Info("✅ WAHA: Successfully sent after session restart")
	return nil
}

// In sendWahaTextMessage, replace the error handling section with:
	if resp.StatusCode == http.StatusInternalServerError {
		bodyStr := string(body)
		if strings.Contains(bodyStr, "websocket not connected") || 
		   strings.Contains(bodyStr, "failed to get device list") ||
		   strings.Contains(bodyStr, "failed to send usync query") {
			logrus.Warn("⚠️ WAHA: Websocket error detected, attempting recovery")
			
			// Create a new request for retry
			retryReq, _ := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
			retryReq.Header.Set("Content-Type", "application/json")
			retryReq.Header.Set("X-Api-Key", apiKey)
			
			if err := h.handleWahaWebsocketError(sessionName, retryReq); err != nil {
				logrus.WithError(err).Error("❌ WAHA: Failed to recover from websocket error")
				return
			}
			return
		}
	}