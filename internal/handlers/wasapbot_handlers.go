package handlers

import (
	"nodepath-chat/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// WasapBotHandlers handles WasapBot related requests
type WasapBotHandlers struct {
	wasapBotRepo repository.WasapBotRepository
}

// NewWasapBotHandlers creates a new WasapBot handlers instance
func NewWasapBotHandlers(wasapBotRepo repository.WasapBotRepository) *WasapBotHandlers {
	return &WasapBotHandlers{
		wasapBotRepo: wasapBotRepo,
	}
}

// GetRepo returns the wasapbot repository for use by other handlers
func (h *WasapBotHandlers) GetRepo() repository.WasapBotRepository {
	return h.wasapBotRepo
}

// GetWasapBotData retrieves WasapBot data with filters
func (h *WasapBotHandlers) GetWasapBotData(c *fiber.Ctx) error {
	// Get user ID from context
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := userIDInterface.(string)

	// Get query parameters
	deviceIDs := c.Query("deviceIds")
	search := c.Query("search")
	status := c.Query("status")
	stage := c.Query("stage")
	dateFrom := c.Query("dateFrom") // Add date from parameter
	dateTo := c.Query("dateTo")     // Add date to parameter
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"device_ids": deviceIDs,
		"search":     search,
		"status":     status,
		"stage":      stage,
		"date_from":  dateFrom,
		"date_to":    dateTo,
		"limit":      limit,
		"offset":     offset,
	}).Info("Getting WasapBot data")

	// Get data from repository with date filters
	data, total, err := h.wasapBotRepo.GetAllWasapBotDataWithDates(limit, offset, deviceIDs, stage, status, search, dateFrom, dateTo, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get WasapBot data")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve data",
		})
	}

	// Log the data being returned
	logrus.WithFields(logrus.Fields{
		"data_count": len(data),
		"total":      total,
		"data":       data,
	}).Info("WasapBot data retrieved")

	// Ensure we return an empty array if no data
	if data == nil {
		data = []map[string]interface{}{}
	}

	// Format response
	response := fiber.Map{
		"records": data,
		"total":   total,
	}

	logrus.WithField("response", response).Info("Sending WasapBot response")

	return c.JSON(response)
}

// GetWasapBotStats retrieves WasapBot statistics
func (h *WasapBotHandlers) GetWasapBotStats(c *fiber.Ctx) error {
	// Get user ID from context
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := userIDInterface.(string)

	// Get device IDs from query
	deviceIDs := c.Query("deviceIds")
	dateFrom := c.Query("dateFrom") // Add date from parameter
	dateTo := c.Query("dateTo")     // Add date to parameter

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"device_ids": deviceIDs,
		"date_from":  dateFrom,
		"date_to":    dateTo,
	}).Info("Getting WasapBot statistics")

	// Get stats from repository with date filters
	stats, err := h.wasapBotRepo.GetWasapBotStatsWithDates(deviceIDs, dateFrom, dateTo, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get WasapBot stats")
		// Return default stats on error
		stats = map[string]interface{}{
			"totalProspects":      0,
			"activeExecutions":    0,
			"completedExecutions": 0,
			"uniqueSchools":       0,
			"uniquePackages":      0,
			"totalWithPhone":      0,
		}
	}

	return c.JSON(stats)
}

// DeleteWasapBotRecord deletes a WasapBot record
func (h *WasapBotHandlers) DeleteWasapBotRecord(c *fiber.Ctx) error {
	// Get user ID from context
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get record ID from params
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid record ID",
		})
	}

	logrus.WithField("id", id).Info("Deleting WasapBot record")

	// Delete the record
	err = h.wasapBotRepo.Delete(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete WasapBot record")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete record",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Record deleted successfully",
	})
}
