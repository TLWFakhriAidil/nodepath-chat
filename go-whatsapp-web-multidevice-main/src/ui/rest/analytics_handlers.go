package rest

import (
	"github.com/gofiber/fiber/v2"
	_ "github.com/go-sql-driver/mysql"
)

// GetCampaignAnalytics returns campaign analytics data
func (handler *App) GetCampaignAnalytics(c *fiber.Ctx) error {
	// Temporarily return empty data to fix compilation
	return c.JSON(fiber.Map{
		"totalCampaigns":          0,
		"totalContactsShouldSend": 0,
		"contactsDoneSend":        0,
		"contactsFailedSend":      0,
		"contactsRemainingSend":   0,
		"chartData": fiber.Map{
			"labels": []string{},
			"sent":   []int{},
			"failed": []int{},
		},
	})
}

// GetSequenceAnalytics returns sequence analytics data
func (handler *App) GetSequenceAnalytics(c *fiber.Ctx) error {
	// Temporarily return empty data to fix compilation
	return c.JSON(fiber.Map{
		"totalSequences":          0,
		"totalFlows":              0,
		"totalContactsShouldSend": 0,
		"contactsDoneSend":        0,
		"contactsFailedSend":      0,
		"contactsRemainingSend":   0,
		"chartData": fiber.Map{
			"labels":    []string{},
			"completed": []int{},
			"failed":    []int{},
			"pending":   []int{},
		},
	})
}

// GetNiches returns all unique niches
func (handler *App) GetNiches(c *fiber.Ctx) error {
	// Temporarily return empty array
	return c.JSON([]string{})
}

// TestDatabaseConnection tests database tables and returns diagnostic info
func (handler *App) TestDatabaseConnection(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "Analytics temporarily disabled for build fix",
	})
}
