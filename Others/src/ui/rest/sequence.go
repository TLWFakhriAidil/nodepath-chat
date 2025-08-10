package rest

import (
	"fmt"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/sequence"
	domainSequence "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/sequence"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type Sequence struct {
	Service sequence.ISequenceUsecase
}

func InitRestSequence(app *fiber.App, service sequence.ISequenceUsecase) {
	rest := Sequence{Service: service}
	
	// Sequence routes
	app.Get("/api/sequences", rest.GetSequences)
	app.Get("/api/sequences/summary", rest.GetSequencesSummary)
	app.Post("/api/sequences", rest.CreateSequence)
	app.Get("/api/sequences/:id", rest.GetSequenceByID)
	app.Put("/api/sequences/:id", rest.UpdateSequence)
	app.Delete("/api/sequences/:id", rest.DeleteSequence)
	
	// Contact management
	app.Post("/api/sequences/:id/contacts", rest.AddContacts)
	app.Get("/api/sequences/:id/contacts", rest.GetContacts)
	app.Delete("/api/sequences/:id/contacts/:contact_id", rest.RemoveContact)
	
	// Actions
	app.Post("/api/sequences/:id/start", rest.StartSequence)
	app.Post("/api/sequences/:id/pause", rest.PauseSequence)
	app.Post("/api/sequences/:id/toggle", rest.ToggleSequence)
	
	// UI routes
	app.Get("/sequences", rest.SequencesPage)
	app.Get("/sequences/:id", rest.SequenceDetailPage)
}

// GetSequences gets all sequences for logged in user
func (controller *Sequence) GetSequences(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}
	
	sequences, err := controller.Service.GetSequences(userID)
	if err != nil {
		logrus.Errorf("Failed to get sequences for user %s: %v", userID, err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	// Count only active sequences
	activeCount := 0
	for _, seq := range sequences {
		if seq.IsActive {
			activeCount++
		}
	}
	logrus.Infof("Found %d active sequences (total: %d) for user %s", activeCount, len(sequences), userID)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequences retrieved",
		Results: sequences,
	})
}

// CreateSequence creates a new sequence
func (controller *Sequence) CreateSequence(c *fiber.Ctx) error {
	var request sequence.CreateSequenceRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	// Set user ID from session
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}
	request.UserID = userID
	
	// Log the request
	logrus.Infof("CreateSequence request: %+v", request)
	logrus.Infof("Number of steps: %d", len(request.Steps))
	
	response, err := controller.Service.CreateSequence(request)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  201,
		Code:    "CREATED",
		Message: "Sequence created successfully",
		Results: response,
	})
}
// GetSequenceByID gets sequence details
func (controller *Sequence) GetSequenceByID(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	response, err := controller.Service.GetSequenceByID(sequenceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Sequence not found",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequence details retrieved",
		Results: response,
	})
}

// UpdateSequence updates a sequence
func (controller *Sequence) UpdateSequence(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	var request sequence.UpdateSequenceRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	err := controller.Service.UpdateSequence(sequenceID, request)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequence updated successfully",
	})
}

// DeleteSequence deletes a sequence
func (controller *Sequence) DeleteSequence(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	err := controller.Service.DeleteSequence(sequenceID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequence deleted successfully",
	})
}
// AddContacts adds contacts to sequence
func (controller *Sequence) AddContacts(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	var request struct {
		Contacts []string `json:"contacts"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	err := controller.Service.AddContactsToSequence(sequenceID, request.Contacts)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Added %d contacts to sequence", len(request.Contacts)),
	})
}

// GetContacts gets all contacts in sequence
func (controller *Sequence) GetContacts(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	contacts, err := controller.Service.GetSequenceContacts(sequenceID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Contacts retrieved",
		Results: contacts,
	})
}

// RemoveContact removes a contact from sequence
func (controller *Sequence) RemoveContact(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	contactID := c.Params("contact_id")
	
	err := controller.Service.RemoveContactFromSequence(sequenceID, contactID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Contact removed from sequence",
	})
}
// StartSequence starts a sequence
func (controller *Sequence) StartSequence(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	err := controller.Service.StartSequence(sequenceID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequence started",
	})
}

// PauseSequence pauses a sequence
func (controller *Sequence) PauseSequence(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	err := controller.Service.PauseSequence(sequenceID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequence paused",
	})
}

// SequencesPage renders sequences page
func (controller *Sequence) SequencesPage(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Redirect("/login")
	}
	
	// Get user's devices
	userRepo := repository.GetUserRepository()
	user, _ := userRepo.GetUserByID(userID)
	devices, _ := userRepo.GetUserDevices(user.ID)
	
	return c.Render("views/sequences", fiber.Map{
		"Title":   "Sequences",
		"User":    user,
		"Devices": devices,
	})
}

// SequenceDetailPage renders sequence detail page
func (controller *Sequence) SequenceDetailPage(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	userID, err := getUserID(c)
	if err != nil {
		return c.Redirect("/login")
	}
	
	// Get user info
	userRepo := repository.GetUserRepository()
	user, _ := userRepo.GetUserByID(userID)
	
	sequence, err := controller.Service.GetSequenceByID(sequenceID)
	if err != nil {
		return c.Redirect("/sequences")
	}
	
	return c.Render("views/sequence_detail", fiber.Map{
		"Title":    "Sequence Detail",
		"Sequence": sequence,
		"User":     user,
	})
}

// GetSequencesSummary gets sequences summary for dashboard
func (controller *Sequence) GetSequencesSummary(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}
	
	// Get user's devices count
	userRepo := repository.GetUserRepository()
	devices, _ := userRepo.GetUserDevices(userID)
	_ = len(devices) // We're not using device count in summary anymore
	
	sequences, err := controller.Service.GetSequences(userID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	// Get all sequence contacts for statistics
	sequenceRepo := repository.GetSequenceRepository()
	
	// Calculate overall summary statistics
	totalFlows := 0
	totalShouldSend := 0
	totalDoneSend := 0
	totalFailedSend := 0
	totalRemainingSend := 0
	
	// Process each sequence to get detailed stats
	detailSequences := make([]interface{}, 0)
	
	for _, seq := range sequences {
		// Get steps for this sequence
		steps, _ := sequenceRepo.GetSequenceSteps(seq.ID)
		flowCount := len(steps)
		totalFlows += flowCount
		
		// Get leads matching this sequence trigger
		// For now, use contacts count as shouldSend
		shouldSend := seq.ContactsCount
		totalShouldSend += shouldSend
		
		// Get sequence contacts for done/failed counts
		contacts, _ := sequenceRepo.GetSequenceContacts(seq.ID)
		doneSend := 0
		failedSend := 0
		
		for _, contact := range contacts {
			if contact.Status == "sent" || contact.Status == "completed" {
				doneSend++
			} else if contact.Status == "failed" {
				failedSend++
			}
		}
		
		totalDoneSend += doneSend
		totalFailedSend += failedSend
		
		remainingSend := shouldSend - doneSend - failedSend
		if remainingSend < 0 {
			remainingSend = 0
		}
		totalRemainingSend += remainingSend
		
		// Add to detail sequences
		detailSequences = append(detailSequences, map[string]interface{}{
			"id":             seq.ID,
			"name":           seq.Name,
			"niche":          seq.Niche,
			"trigger":        seq.Trigger,
			"status":         seq.Status,
			"total_flows":    flowCount,
			"should_send":    shouldSend,
			"done_send":      doneSend,
			"failed_send":    failedSend,
			"remaining_send": remainingSend,
		})
	}
	
	// Calculate summary
	summary := map[string]interface{}{
		"sequences": map[string]int{
			"total":    len(sequences),
			"active":   0,
			"inactive": 0,
		},
		"total_flows":          totalFlows,
		"total_should_send":    totalShouldSend,
		"total_done_send":      totalDoneSend,
		"total_failed_send":    totalFailedSend,
		"total_remaining_send": totalRemainingSend,
		"recent_sequences":     detailSequences,
	}
	
	// Count active/inactive
	for _, seq := range sequences {
		if seq.Status == "active" {
			summary["sequences"].(map[string]int)["active"]++
		} else {
			summary["sequences"].(map[string]int)["inactive"]++
		}
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Sequences summary retrieved",
		Results: summary,
	})
}

// ToggleSequence toggles sequence status between active and inactive
func (controller *Sequence) ToggleSequence(c *fiber.Ctx) error {
	sequenceID := c.Params("id")
	
	// Get current sequence
	sequence, err := controller.Service.GetSequenceByID(sequenceID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Sequence not found",
		})
	}
	
	// Toggle status
	newStatus := "inactive"
	if sequence.Status != "active" {
		newStatus = "active"
	}
	
	// Update sequence
	updateReq := domainSequence.UpdateSequenceRequest{
		Status:   newStatus,
		IsActive: newStatus == "active",
	}
	
	err = controller.Service.UpdateSequence(sequenceID, updateReq)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: err.Error(),
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Sequence %s successfully", newStatus),
		Results: map[string]string{
			"status": newStatus,
		},
	})
}