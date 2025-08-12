package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/database"
	"nodepath-chat/internal/handlers"
	"nodepath-chat/internal/services"
	"nodepath-chat/internal/whatsapp"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		logrus.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}
	logrus.Info("Database initialized successfully")

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logrus.WithError(err).Fatal("Failed to run migrations")
	}
	logrus.Info("Database migrations completed")

	// Initialize Redis
	redisClient := services.InitializeRedis(cfg)
	logrus.Info("Redis initialized successfully")

	// Initialize services
	flowService := services.NewFlowService(db, redisClient)
	chatService := services.NewChatService(db, redisClient)
	aiService := services.NewAIService(cfg)
	queueService := services.NewQueueService(redisClient)
	deviceSettingsService := services.NewDeviceSettingsService(db)

	whatsappService, err := whatsapp.NewService(cfg, chatService, queueService)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize WhatsApp service")
	}

	// Initialize handlers
	handlers := handlers.NewHandlers(
		flowService,
		chatService,
		aiService,
		queueService,
		whatsappService,
		deviceSettingsService,
	)

	// Initialize HTML template engine
	engine := html.New("./templates", ".html")
	engine.Reload(cfg.AppEnv == "development")
	
	// Add template functions
	engine.AddFunc("now", func() time.Time {
		return time.Now()
	})

	// Create Fiber app
	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: customErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10MB
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	if cfg.AppEnv == "development" {
		app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
		}))
	}

	// Static files for React app
	app.Static("/", "./dist")
	app.Static("/static", "./static") // Keep for backward compatibility

	// Health check endpoint
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// Setup API routes
	api := app.Group("/api")
	handlers.SetupRoutes(api)

	// Catch-all route for React Router (SPA)
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./dist/index.html")
	})

	// Start background services
	go whatsappService.StartQueueProcessor()
	go func() {
		for {
			if err := queueService.ProcessDelayedMessages(); err != nil {
				logrus.WithError(err).Error("Error processing delayed messages")
			}
			time.Sleep(30 * time.Second)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logrus.Info("Shutting down server...")
		whatsappService.Disconnect()
		app.Shutdown()
	}()

	// Start server
	logrus.Infof("Server starting on port %d", cfg.Port)
	if err := app.Listen(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		logrus.WithError(err).Fatal("Failed to start server")
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Log error
	logrus.Errorf("Error %d: %v", code, err)

	// Return JSON error for API routes
	if c.Path() != "" && len(c.Path()) >= 4 && c.Path()[:4] == "/api" {
		return c.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
			"code":    code,
		})
	}

	// Return error page for web routes
	return c.Status(code).Render("error", fiber.Map{
		"Title":   fmt.Sprintf("Error %d", code),
		"Code":    code,
		"Message": err.Error(),
	})
}