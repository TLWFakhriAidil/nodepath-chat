package handlers

import (
	"database/sql"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type StageSetValueHandlers struct {
	service *services.StageSetValueService
}

func NewStageSetValueHandlers(db *sql.DB) *StageSetValueHandlers {
	return &StageSetValueHandlers{
		service: services.NewStageSetValueService(db),
	}
}

// GetAll handles GET /api/set-stage
func (h *StageSetValueHandlers) GetAll(c *fiber.Ctx) error {
	values, err := h.service.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(values)
}

// GetByDevice handles GET /api/set-stage/device/:deviceId
func (h *StageSetValueHandlers) GetByDevice(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "device ID is required",
		})
	}

	values, err := h.service.GetByDeviceID(deviceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(values)
}

// Create handles POST /api/set-stage
func (h *StageSetValueHandlers) Create(c *fiber.Ctx) error {
	var value models.StageSetValue

	if err := c.BodyParser(&value); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Handle inputHardCode field
	if value.TypeInputData == "Set" {
		// Parse the inputHardCode from JSON
		var req struct {
			IDDevice      string  `json:"id_device"`
			Stage         int     `json:"stage"`
			TypeInputData string  `json:"type_inputData"`
			ColumnsData   string  `json:"columnsData"`
			InputHardCode *string `json:"inputHardCode"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		value.IDDevice = req.IDDevice
		value.Stage = req.Stage
		value.TypeInputData = req.TypeInputData
		value.ColumnsData = req.ColumnsData

		if req.InputHardCode != nil && *req.InputHardCode != "" {
			value.InputHardCode = sql.NullString{
				String: *req.InputHardCode,
				Valid:  true,
			}
		}
	}

	if err := h.service.Create(&value); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(value)
}

// Update handles PUT /api/set-stage/:id
func (h *StageSetValueHandlers) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid ID",
		})
	}

	var value models.StageSetValue
	if err := c.BodyParser(&value); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	value.StageSetValueID = id

	// Handle inputHardCode field
	if value.TypeInputData == "Set" {
		var req struct {
			Stage         int     `json:"stage"`
			TypeInputData string  `json:"type_inputData"`
			ColumnsData   string  `json:"columnsData"`
			InputHardCode *string `json:"inputHardCode"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		value.Stage = req.Stage
		value.TypeInputData = req.TypeInputData
		value.ColumnsData = req.ColumnsData

		if req.InputHardCode != nil && *req.InputHardCode != "" {
			value.InputHardCode = sql.NullString{
				String: *req.InputHardCode,
				Valid:  true,
			}
		}
	}

	if err := h.service.Update(&value); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(value)
}

// Delete handles DELETE /api/set-stage/:id
func (h *StageSetValueHandlers) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid ID",
		})
	}

	if err := h.service.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
