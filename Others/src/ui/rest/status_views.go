package rest

import (
	"github.com/gofiber/fiber/v2"
)

// RedisStatusView renders the Redis status page
func (rest *App) RedisStatusView(c *fiber.Ctx) error {
	return c.Render("views/redis_status", fiber.Map{
		"Title": "Redis Status",
	})
}

// DeviceWorkerStatusView renders the device worker status page
func (rest *App) DeviceWorkerStatusView(c *fiber.Ctx) error {
	return c.Render("views/device_worker_status", fiber.Map{
		"Title": "Device Worker Status",
	})
}

// AllWorkersStatusView renders the all workers status page
func (rest *App) AllWorkersStatusView(c *fiber.Ctx) error {
	return c.Render("views/all_workers_status", fiber.Map{
		"Title": "All Workers Status",
	})
}
