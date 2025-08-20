package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/database"
	"nodepath-chat/internal/handlers"
	"nodepath-chat/internal/repository"
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

	// Initialize database (skip if MYSQL_URI is empty or connection fails)
	var db *sql.DB
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		logrus.Warn("MYSQL_URI is empty, running without database")
	} else {
		var err error
		db, err = database.Initialize(cfg)
		if err != nil {
			logrus.WithError(err).Warn("Failed to initialize database, continuing without database")
			db = nil
		} else {
			logrus.Info("Database initialized successfully")
			
			// Run migrations
			if err := database.RunMigrations(db); err != nil {
				logrus.WithError(err).Warn("Failed to run migrations, continuing anyway")
			} else {
				logrus.Info("Database migrations completed")
			}
		}
	}

	// Initialize Redis with clustering support
	redisClient := services.InitializeRedis(cfg)
	logrus.Info("Redis initialized successfully")

	// Initialize performance-optimized services
	// Handle Redis client for services that need concrete type
	var concreteRedisClient *redis.Client
	if redisClient != nil {
		var ok bool
		concreteRedisClient, ok = redisClient.(*redis.Client)
		if !ok {
			logrus.Warn("Redis client type assertion failed, using nil client")
			concreteRedisClient = nil
		}
	} else {
		logrus.Warn("Redis not available, services will run without caching")
	}
	
	flowService := services.NewFlowService(db, concreteRedisClient)
	chatService := services.NewChatService(db, concreteRedisClient)
	aiService := services.NewAIService(cfg)
	queueService := services.NewQueueService(redisClient)
	deviceSettingsService := services.NewDeviceSettingsService(db)
	
	// Initialize repositories
	aiWhatsappRepo := repository.NewAIWhatsappRepository(db)
	deviceSettingsRepo := repository.NewDeviceSettingsRepository(db)
	logrus.Info("Repositories initialized successfully")
	
	// Initialize AI WhatsApp service
	aiWhatsappService := services.NewAIWhatsappService(aiWhatsappRepo, deviceSettingsRepo, flowService)
	logrus.Info("AI WhatsApp service initialized successfully")
	
	// Initialize WebSocket service for real-time communication
	websocketService := services.NewWebSocketService(cfg.MaxConcurrentUsers)
	logrus.Info("WebSocket service initialized for real-time messaging")
	
	// Initialize media service with CDN support
	mediaService := services.NewMediaService(cfg.CDNEnabled, cfg.CDNBaseURL, "./media")
	logrus.Info("Media service initialized with CDN support")

	// Initialize provider service for message sending
	providerService := services.NewProviderService()
	logrus.Info("Provider service initialized for Wablas/Whacenter APIs")

	// Initialize WhatsApp service with multi-device support
	logrus.Info("🔧 MAIN: About to initialize WhatsApp service...")
	logrus.Info("🔧 MAIN: Initializing WhatsApp service...")
	whatsappService, err := whatsapp.NewService(cfg, chatService, queueService, flowService, aiService, aiWhatsappService, websocketService, deviceSettingsService, providerService)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize WhatsApp service")
	}
	logrus.Info("✅ MAIN: WhatsApp service initialized successfully")

	// Initialize handlers with all services
	handlers := handlers.NewHandlers(
		flowService,
		chatService,
		aiService,
		queueService,
		whatsappService,
		deviceSettingsService,
		websocketService,
		mediaService,
		db,
	)

	// Initialize HTML template engine
	engine := html.New("./templates", ".html")
	engine.Reload(cfg.AppEnv == "development")
	
	// Add template functions
	engine.AddFunc("now", func() time.Time {
		return time.Now()
	})

	// Create Fiber app with performance optimizations
	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: customErrorHandler,
		BodyLimit:    50 * 1024 * 1024, // 50MB for media files
		ReadTimeout:  30 * time.Second,  // Increased for large uploads
		WriteTimeout: 30 * time.Second,  // Increased for large downloads
		IdleTimeout:  120 * time.Second, // Keep connections alive longer
		Concurrency:  cfg.MaxConcurrentUsers * 2, // Handle high concurrency
	})

	// Performance and security middleware
	app.Use(recover.New())
	
	// Rate limiting for API protection
	app.Use(limiter.New(limiter.Config{
		Max:        100, // 100 requests per minute per IP
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit by IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	}))
	
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Device-ID",
		AllowCredentials: false, // Set to false when using wildcard origins
	}))

	if cfg.AppEnv == "development" {
		app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} (${latency}) - ${ip}\n",
		}))
	}

	// Health check endpoint with performance metrics
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":              "ok",
			"time":                time.Now().Unix(),
			"websocket_connections": websocketService.GetConnectionCount(),
			"max_concurrent_users": cfg.MaxConcurrentUsers,
			"cdn_enabled":          cfg.CDNEnabled,
		})
	})

	// WebSocket endpoint for real-time communication
	app.Use("/ws", func(c *fiber.Ctx) error {
		// Check if connection is a WebSocket upgrade
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocketService.HandleWebSocket)

	// Media endpoints for file upload and serving
	media := app.Group("/media")
	media.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No file uploaded",
			})
		}
		
		result, err := mediaService.UploadFile(file)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		
		return c.JSON(result)
	})
	
	media.Get("/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		data, mimeType, err := mediaService.ServeFile(filename)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "File not found",
			})
		}
		
		c.Set("Content-Type", mimeType)
		c.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		return c.Send(data)
	})
	
	media.Get("/thumbnails/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		data, mimeType, err := mediaService.ServeFile("thumbnails/" + filename)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Thumbnail not found",
			})
		}
		
		c.Set("Content-Type", mimeType)
		c.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		return c.Send(data)
	})

	// Setup API routes
	api := app.Group("/api")
	handlers.SetupRoutes(api)

	// Static files for React app (after API routes)
	app.Static("/", "./dist")
	app.Static("/static", "./static") // Keep for backward compatibility

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