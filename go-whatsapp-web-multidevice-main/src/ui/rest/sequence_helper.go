package rest

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
)

// Package level helper to get userID safely
func getUserID(c *fiber.Ctx) (string, error) {
	// First try to get from locals
	userID := c.Locals("userID")
	if userID != nil {
		if id, ok := userID.(string); ok {
			return id, nil
		}
	}
	
	// Fallback: try to get from email in locals
	email := c.Locals("email")
	if email != nil {
		if emailStr, ok := email.(string); ok {
			userRepo := repository.GetUserRepository()
			user, err := userRepo.GetUserByEmail(emailStr)
			if err == nil {
				return user.ID, nil
			}
		}
	}
	
	// Last resort: try session cookie
	token := c.Cookies("session_token")
	if token != "" {
		userRepo := repository.GetUserRepository()
		session, err := userRepo.GetSession(token)
		if err == nil {
			return session.UserID, nil
		}
	}
	
	return "", fmt.Errorf("user not authenticated")
}
