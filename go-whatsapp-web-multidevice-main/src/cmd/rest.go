package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/helpers"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/middleware"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/websocket"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/usecase"
	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Send whatsapp API over http",
	Long:  `This application is from clone https://github.com/aldinokemal/go-whatsapp-web-multidevice`,
	Run:   restServer,
}

func init() {
	rootCmd.AddCommand(restCmd)
}
func restServer(_ *cobra.Command, _ []string) {
	engine := html.NewFileSystem(http.FS(EmbedViews), ".html")
	engine.AddFunc("isEnableBasicAuth", func(token any) bool {
		return token != nil
	})
	app := fiber.New(fiber.Config{
		Views:     engine,
		BodyLimit: int(config.WhatsappSettingMaxVideoSize),
		Network:   "tcp",
	})

	app.Static("/statics", "./statics")
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendFile("./statics/favicon.ico")
	})
	app.Use("/components", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/components",
		Browse:     true,
	}))
	app.Use("/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/assets",
		Browse:     true,
	}))

	app.Use(middleware.Recovery())
	
	// IMPORTANT: Register public routes BEFORE auth middleware
	rest.InitPublicDeviceRoutes(app, database.GetDB()) // Add public device view routes
	rest.InitPublicDeviceAPI(app) // Add public device API endpoints
	
	// Now apply auth middleware - it won't affect routes registered above
	app.Use(middleware.BasicAuth())
	app.Use(middleware.CustomAuth()) // Add custom auth middleware
	if config.AppDebug {
		app.Use(logger.New())
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Auth-Token",
	}))

	// Comment out basic auth to use custom login
	/*
	if len(config.AppBasicAuthCredential) > 0 {
		account := make(map[string]string)
		for _, basicAuth := range config.AppBasicAuthCredential {
			ba := strings.Split(basicAuth, ":")
			if len(ba) != 2 {
				log.Fatalln("Basic auth is not valid, please this following format <user>:<secret>")
			}
			account[ba[0]] = ba[1]
		}

		app.Use(basicauth.New(basicauth.Config{
			Users: account,
		}))
	}
	*/

	// Rest
	appRest := rest.InitRestApp(app, appUsecase)
	appRest.SetSendService(sendUsecase) // Set send service for WhatsApp Web
	rest.InitRestSend(app, sendUsecase)
	rest.InitRestUser(app, userUsecase)
	rest.InitRestMessage(app, messageUsecase)
	rest.InitRestGroup(app, groupUsecase)
	rest.InitRestNewsletter(app, newsletterUsecase)
	rest.InitRestSequence(app, sequenceUsecase)
	rest.InitRestMonitoring(app) // Add monitoring endpoints
	rest.InitWorkerControlAPI(app) // Add worker control endpoints
	rest.InitTeamRoutes(app, database.GetDB()) // Add team member routes
	rest.InitRedisCleanupAPI(app) // Add Redis cleanup endpoints
	rest.InitWebhookLead(app) // Add webhook endpoint for creating leads

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("views/index", fiber.Map{
			"AppHost":        fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname()),
			"AppVersion":     config.AppVersion,
			"BasicAuthToken": c.UserContext().Value(middleware.AuthorizationValue("BASIC_AUTH")),
			"MaxFileSize":    humanize.Bytes(uint64(config.WhatsappSettingMaxFileSize)),
			"MaxVideoSize":   humanize.Bytes(uint64(config.WhatsappSettingMaxVideoSize)),
		})
	})
	
	// Redis cleanup page
	app.Get("/redis-cleanup", middleware.BasicAuth(), func(c *fiber.Ctx) error {
		return c.Render("views/redis_cleanup", fiber.Map{})
	})

	websocket.RegisterRoutes(app, appUsecase)
	go websocket.RunHub()

	// REMOVED: Auto-reconnect is now manual via Refresh button
	// whatsapp.StartMultiDeviceAutoReconnect()
	
	// Start auto flush chat csv
	if config.WhatsappChatStorage {
		go helpers.StartAutoFlushChatStorage()
	}
	
	// Start broadcast manager
	_ = broadcast.GetBroadcastManager()
	logrus.Info("Broadcast manager started")
	
	// Optimize system for 3000 devices
	broadcast.OptimizeFor3000Devices()
	
	// Load all existing WhatsApp sessions on startup
	go func() {
		logrus.Info("Waiting 5 seconds before loading WhatsApp devices...")
		time.Sleep(5 * time.Second)
		whatsapp.LoadAllDevicesOnStartup()
	}()
	
	// DISABLED - Using self-healing per message instead
	// healthMonitor := whatsapp.GetDeviceHealthMonitor(whatsappDB)
	// healthMonitor.Start()
	logrus.Info("🔄 SELF-HEALING MODE: Workers refresh clients per message (no background keepalive)")
	
	// Start the ultra-optimized broadcast processor for 3000+ devices
	// This processor creates broadcast-specific worker pools
	go usecase.StartUltraOptimizedBroadcastProcessor()
	logrus.Info("Ultra-optimized broadcast processor started (3000+ device support)")
	
	// Start campaign trigger processor using optimized version
	go func() {
		db := database.GetDB()
		campaignTrigger := usecase.NewOptimizedCampaignTrigger(db)
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		logrus.Info("Campaign trigger processor started (checks every minute)")
		
		// Process immediately on start
		campaignTrigger.ProcessCampaigns()
		
		// Then process every minute
		for range ticker.C {
			campaignTrigger.ProcessCampaigns()
		}
	}()
	
	// Start sequence trigger processor for trigger-based flow
	go usecase.StartSequenceTriggerProcessor()
	logrus.Info("Sequence trigger processor started")
	
	// Start campaign status monitor
	go usecase.StartCampaignStatusMonitor()
	logrus.Info("Campaign status monitor started")
	
	// Start queued message cleaner
	go usecase.StartQueuedMessageCleaner()
	logrus.Info("Queued message cleaner started")
	
	// Start broadcast coordinator
	go usecase.StartBroadcastCoordinator()
	logrus.Info("Broadcast coordinator started")
	
	// Start broadcast worker processor - CRITICAL FOR WORKER POOL
	go usecase.StartBroadcastWorkerProcessor()
	
	// Start cleanup worker for stuck messages
	go repository.StartCleanupWorker()
	logrus.Info("Broadcast worker processor started - using Worker Pool System")
	
	// Start campaign completion checker
	go usecase.StartCampaignCompletionChecker()
	logrus.Info("Campaign completion checker started")
	
	// Auto-reconnect devices on startup - DISABLED
	// Using MonitorDeviceErrors instead for continuous monitoring
	/*
	go func() {
		time.Sleep(5 * time.Second) // Wait a bit for services to initialize
		logrus.Info("Starting auto-reconnect for online devices...")
		whatsapp.AutoReconnectOnlineDevices()
	}()
	*/

	if err := app.Listen(":" + config.AppPort); err != nil {
		log.Fatalln("Failed to start: ", err.Error())
	}
}