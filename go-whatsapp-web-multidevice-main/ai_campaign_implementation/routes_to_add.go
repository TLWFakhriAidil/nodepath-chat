// Add these routes to cmd/rest.go in the appropriate section

// AI Lead Management Routes
app.Post("/api/leads-ai", handler.CreateLeadAI)
app.Get("/api/leads-ai", handler.GetLeadsAI)
app.Put("/api/leads-ai/:id", handler.UpdateLeadAI)
app.Delete("/api/leads-ai/:id", handler.DeleteLeadAI)

// AI Campaign Trigger Route
app.Post("/api/campaigns-ai/:id/trigger", handler.TriggerAICampaign)

// Note: The regular campaign creation endpoint (/api/campaigns) 
// will be modified to support AI campaigns by checking for the 'ai' field