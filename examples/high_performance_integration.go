package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nodepath-chat/internal/handlers"
	"nodepath-chat/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HighPerformanceApp demonstrates how to integrate all high-performance components
type HighPerformanceApp struct {
	// Core services (your existing services)
	chatService     *services.ChatService
	flowService     *services.FlowService
	aiService       *services.AIService
	queueService    *services.QueueService
	whatsappService *services.WhatsappService
	
	// High-performance components
	hpManager       *services.HighPerformanceManager
	
	// HTTP server
	server          *http.Server
	router          *gin.Engine
}

// NewHighPerformanceApp creates a new high-performance application
func NewHighPerformanceApp() *HighPerformanceApp {
	// Initialize your existing services here
	// These are placeholders - replace with your actual service initialization
	chatService := &services.ChatService{} // Initialize with your actual chat service
	flowService := &services.FlowService{} // Initialize with your actual flow service
	aiService := &services.AIService{}     // Initialize with your actual AI service
	queueService := &services.QueueService{} // Initialize with your actual queue service
	whatsappService := &services.WhatsappService{} // Initialize with your actual WhatsApp service
	
	// Create high-performance manager
	hpManager := services.NewHighPerformanceManager(
		chatService,
		flowService,
		aiService,
		queueService,
		whatsappService,
	)
	
	// Setup Gin router with optimizations
	router := setupOptimizedRouter()
	
	// Register high-performance routes
	hpManager.RegisterRoutes(router)
	
	// Create HTTP server with optimizations
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	
	return &HighPerformanceApp{
		chatService:     chatService,
		flowService:     flowService,
		aiService:       aiService,
		queueService:    queueService,
		whatsappService: whatsappService,
		hpManager:       hpManager,
		server:          server,
		router:          router,
	}
}

// Start starts the high-performance application
func (app *HighPerformanceApp) Start() error {
	logrus.Info("Starting high-performance WhatsApp chat system...")
	
	// Start high-performance manager
	if err := app.hpManager.Start(); err != nil {
		return err
	}
	
	// Start HTTP server
	go func() {
		logrus.WithField("addr", app.server.Addr).Info("Starting HTTP server")
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()
	
	// Log system status
	go app.logSystemStatus()
	
	logrus.Info("High-performance system started successfully")
	return nil
}

// Stop stops the application gracefully
func (app *HighPerformanceApp) Stop() error {
	logrus.Info("Shutting down high-performance system...")
	
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Shutdown HTTP server
	if err := app.server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Error shutting down HTTP server")
	}
	
	// Stop high-performance manager
	if err := app.hpManager.Stop(); err != nil {
		logrus.WithError(err).Error("Error stopping high-performance manager")
	}
	
	logrus.Info("High-performance system shutdown complete")
	return nil
}

// setupOptimizedRouter sets up Gin router with performance optimizations
func setupOptimizedRouter() *gin.Engine {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	
	// Add middleware for performance
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(compressionMiddleware())
	router.Use(rateLimitMiddleware())
	router.Use(metricsMiddleware())
	
	// Disable trusted proxies for security
	router.SetTrustedProxies(nil)
	
	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// compressionMiddleware adds response compression
func compressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple compression middleware
		// In production, use a proper compression middleware like gin-gzip
		c.Header("Content-Encoding", "gzip")
		c.Next()
	}
}

// rateLimitMiddleware adds basic rate limiting
func rateLimitMiddleware() gin.HandlerFunc {
	// Create a simple rate limiter
	rateLimiter := services.NewRateLimiter(1000) // 1000 requests per second
	
	return func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": 1,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// metricsMiddleware adds request metrics
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		// Log request metrics
		duration := time.Since(start)
		logrus.WithFields(logrus.Fields{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      c.Writer.Status(),
			"duration":    duration,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
		}).Debug("HTTP request processed")
	}
}

// logSystemStatus periodically logs system status
func (app *HighPerformanceApp) logSystemStatus() {
	ticker := time.NewTicker(60 * time.Second) // Log every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			status := app.hpManager.GetSystemStatus()
			logrus.WithFields(logrus.Fields{
				"concurrent_users":     status["system_metrics"].(map[string]interface{})["concurrent_users"],
				"messages_per_second":  status["system_metrics"].(map[string]interface{})["messages_per_second"],
				"average_response_time": status["system_metrics"].(map[string]interface{})["average_response_time"],
				"memory_usage":         status["system_metrics"].(map[string]interface{})["memory_usage"],
				"goroutine_count":      status["system_metrics"].(map[string]interface{})["goroutine_count"],
			}).Info("System status report")
		}
	}
}

// Example of how to process a WhatsApp message using the high-performance system
func (app *HighPerformanceApp) ProcessWhatsAppMessage(deviceID, phoneNumber, content string) error {
	// Create incoming message
	msg := &services.IncomingMessage{
		ID:          generateMessageID(deviceID, phoneNumber),
		Type:        services.MessageTypeUserReply,
		DeviceID:    deviceID,
		PhoneNumber: phoneNumber,
		Content:     content,
		Timestamp:   time.Now(),
		Priority:    3, // Normal priority
		MaxRetries:  3,
	}
	
	// Process through high-performance manager
	return app.hpManager.ProcessMessage(msg)
}

// generateMessageID generates a unique message ID
func generateMessageID(deviceID, phoneNumber string) string {
	return fmt.Sprintf("%s_%s_%d", deviceID, phoneNumber, time.Now().UnixNano())
}

// main function demonstrates how to run the high-performance system
func main() {
	// Setup logging
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	
	// Create and start the application
	app := NewHighPerformanceApp()
	
	// Start the application
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
	
	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logrus.Info("Shutdown signal received")
	
	// Graceful shutdown
	if err := app.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}

// Example usage functions

// ExampleProcessUserMessage shows how to process a user message
func ExampleProcessUserMessage() {
	app := NewHighPerformanceApp()
	app.Start()
	
	// Process a user message
	err := app.ProcessWhatsAppMessage("device123", "+1234567890", "Hello, I need help!")
	if err != nil {
		logrus.WithError(err).Error("Failed to process message")
	}
}

// ExampleGetSystemStatus shows how to get system status
func ExampleGetSystemStatus() {
	app := NewHighPerformanceApp()
	app.Start()
	
	// Get system status
	status := app.hpManager.GetSystemStatus()
	logrus.WithField("status", status).Info("Current system status")
}

// ExamplePerformanceReport shows how to get a performance report
func ExamplePerformanceReport() {
	app := NewHighPerformanceApp()
	app.Start()
	
	// Get performance report
	report := app.hpManager.GetPerformanceReport()
	logrus.WithField("report", report).Info("Performance report")
}

// Performance testing function
func PerformanceTest() {
	app := NewHighPerformanceApp()
	app.Start()
	
	// Simulate high load
	logrus.Info("Starting performance test...")
	
	// Process 1000 messages concurrently
	for i := 0; i < 1000; i++ {
		go func(id int) {
			msg := &services.IncomingMessage{
				ID:          fmt.Sprintf("test_msg_%d", id),
				Type:        services.MessageTypeUserReply,
				DeviceID:    "test_device",
				PhoneNumber: fmt.Sprintf("+123456789%d", id),
				Content:     fmt.Sprintf("Test message %d", id),
				Timestamp:   time.Now(),
				Priority:    3,
				MaxRetries:  3,
			}
			
			if err := app.hpManager.ProcessMessage(msg); err != nil {
				logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to process test message")
			}
		}(i)
	}
	
	// Wait and check performance
	time.Sleep(10 * time.Second)
	report := app.hpManager.GetPerformanceReport()
	logrus.WithField("performance_report", report).Info("Performance test completed")
}