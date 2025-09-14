package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// TEMP FILE - Extract the validateWebhookPayload function properly
func (h *Handlers) validateWebhookPayload(data map[string]interface{}) error {
	// Check payload size limit (1MB)
	payloadBytes, _ := json.Marshal(data)
	if len(payloadBytes) > 1024*1024 {
		return fmt.Errorf("payload too large: %d bytes", len(payloadBytes))
	}
	
	// Check for potentially malicious content
	for key, value := range data {
		if strValue, ok := value.(string); ok {
			// Check for common injection patterns
			if err := h.checkForInjection(key, strValue); err != nil {
				return err
			}
		}
	}
	
	return nil
}
