package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name         string                 `json:"name"`
	Status       HealthStatus           `json:"status"`
	Message      string                 `json:"message"`
	LastChecked  time.Time              `json:"last_checked"`
	ResponseTime time.Duration          `json:"response_time"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status     HealthStatus                `json:"status"`
	Timestamp  time.Time                   `json:"timestamp"`
	Uptime     time.Duration               `json:"uptime"`
	Version    string                      `json:"version"`
	Components map[string]*ComponentHealth `json:"components"`
}

// HealthService provides comprehensive health checks for all system components
type HealthService struct {
	db           *sql.DB
	redis        *redis.Client
	startTime    time.Time
	version      string
	mu           sync.RWMutex
	lastCheck    time.Time
	cachedHealth *SystemHealth
	cacheTimeout time.Duration
}

// NewHealthService creates a new health service
func NewHealthService(db *sql.DB, redis *redis.Client, version string) *HealthService {
	return &HealthService{
		db:           db,
		redis:        redis,
		startTime:    time.Now(),
		version:      version,
		cacheTimeout: 30 * time.Second, // Cache health checks for 30 seconds
	}
}

// GetSystemHealth returns comprehensive system health status
func (h *HealthService) GetSystemHealth(ctx context.Context) *SystemHealth {
	h.mu.RLock()
	if h.cachedHealth != nil && time.Since(h.lastCheck) < h.cacheTimeout {
		defer h.mu.RUnlock()
		return h.cachedHealth
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check after acquiring write lock
	if h.cachedHealth != nil && time.Since(h.lastCheck) < h.cacheTimeout {
		return h.cachedHealth
	}

	logrus.Debug("Performing comprehensive system health check")

	health := &SystemHealth{
		Timestamp:  time.Now(),
		Uptime:     time.Since(h.startTime),
		Version:    h.version,
		Components: make(map[string]*ComponentHealth),
	}

	// Check all components concurrently for better performance
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Database health check
	wg.Add(1)
	go func() {
		defer wg.Done()
		dbHealth := h.checkDatabaseHealth(ctx)
		mu.Lock()
		health.Components["database"] = dbHealth
		mu.Unlock()
	}()

	// Redis health check
	wg.Add(1)
	go func() {
		defer wg.Done()
		redisHealth := h.checkRedisHealth(ctx)
		mu.Lock()
		health.Components["redis"] = redisHealth
		mu.Unlock()
	}()

	// Memory health check
	wg.Add(1)
	go func() {
		defer wg.Done()
		memoryHealth := h.checkMemoryHealth()
		mu.Lock()
		health.Components["memory"] = memoryHealth
		mu.Unlock()
	}()

	// Disk health check
	wg.Add(1)
	go func() {
		defer wg.Done()
		diskHealth := h.checkDiskHealth()
		mu.Lock()
		health.Components["disk"] = diskHealth
		mu.Unlock()
	}()

	// External API health check
	wg.Add(1)
	go func() {
		defer wg.Done()
		apiHealth := h.checkExternalAPIHealth(ctx)
		mu.Lock()
		health.Components["external_apis"] = apiHealth
		mu.Unlock()
	}()

	wg.Wait()

	// Determine overall system health
	health.Status = h.calculateOverallHealth(health.Components)

	// Cache the result
	h.cachedHealth = health
	h.lastCheck = time.Now()

	logrus.WithFields(logrus.Fields{
		"overall_status": health.Status,
		"uptime":         health.Uptime,
		"components":     len(health.Components),
	}).Info("System health check completed")

	return health
}

// checkDatabaseHealth performs comprehensive database health checks
func (h *HealthService) checkDatabaseHealth(ctx context.Context) *ComponentHealth {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "database",
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	if h.db == nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Database connection not initialized"
		health.ResponseTime = time.Since(start)
		return health
	}

	// Test basic connectivity
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctxWithTimeout); err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("Database ping failed: %v", err)
		health.ResponseTime = time.Since(start)
		return health
	}

	// Check database version and basic info
	var version, database string
	var connections int

	if err := h.db.QueryRowContext(ctxWithTimeout, "SELECT VERSION()").Scan(&version); err != nil {
		health.Status = HealthStatusDegraded
		health.Message = fmt.Sprintf("Failed to get database version: %v", err)
	} else {
		health.Details["version"] = version
	}

	if err := h.db.QueryRowContext(ctxWithTimeout, "SELECT DATABASE()").Scan(&database); err != nil {
		health.Status = HealthStatusDegraded
		health.Message = fmt.Sprintf("Failed to get database name: %v", err)
	} else {
		health.Details["database"] = database
	}

	// Check active connections
	if err := h.db.QueryRowContext(ctxWithTimeout, "SHOW STATUS LIKE 'Threads_connected'").Scan(nil, &connections); err != nil {
		logrus.WithError(err).Debug("Failed to get connection count")
	} else {
		health.Details["active_connections"] = connections
	}

	// Test a simple query on a known table
	var count int
	if err := h.db.QueryRowContext(ctxWithTimeout, "SELECT COUNT(*) FROM device_setting_nodepath").Scan(&count); err != nil {
		health.Status = HealthStatusDegraded
		health.Message = fmt.Sprintf("Failed to query device settings table: %v", err)
	} else {
		health.Details["device_settings_count"] = count
	}

	if health.Status == "" {
		health.Status = HealthStatusHealthy
		health.Message = "Database is healthy"
	}

	health.ResponseTime = time.Since(start)
	return health
}

// checkRedisHealth performs Redis health checks
func (h *HealthService) checkRedisHealth(ctx context.Context) *ComponentHealth {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "redis",
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	if h.redis == nil {
		health.Status = HealthStatusDegraded
		health.Message = "Redis client not initialized (optional component)"
		health.ResponseTime = time.Since(start)
		return health
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Test basic connectivity
	if err := h.redis.Ping(ctxWithTimeout).Err(); err != nil {
		health.Status = HealthStatusDegraded
		health.Message = fmt.Sprintf("Redis ping failed: %v", err)
		health.ResponseTime = time.Since(start)
		return health
	}

	// Get Redis info
	if info, err := h.redis.Info(ctxWithTimeout).Result(); err != nil {
		health.Status = HealthStatusDegraded
		health.Message = fmt.Sprintf("Failed to get Redis info: %v", err)
	} else {
		// Parse basic info from Redis INFO command
		health.Details["info_available"] = len(info) > 0
	}

	if health.Status == "" {
		health.Status = HealthStatusHealthy
		health.Message = "Redis is healthy"
	}

	health.ResponseTime = time.Since(start)
	return health
}

// checkMemoryHealth performs memory usage health checks
func (h *HealthService) checkMemoryHealth() *ComponentHealth {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "memory",
		LastChecked: start,
		Details:     make(map[string]interface{}),
		Status:      HealthStatusHealthy,
		Message:     "Memory monitoring not implemented for Windows",
	}

	// Note: Memory monitoring would require platform-specific implementation
	// For Windows, we would need to use Windows APIs or external tools
	// This is a placeholder for future implementation

	health.ResponseTime = time.Since(start)
	return health
}

// checkDiskHealth performs disk space health checks
func (h *HealthService) checkDiskHealth() *ComponentHealth {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "disk",
		LastChecked: start,
		Details:     make(map[string]interface{}),
		Status:      HealthStatusHealthy,
		Message:     "Disk monitoring not implemented for Windows",
	}

	// Note: Disk monitoring would require platform-specific implementation
	// For Windows, we would need to use Windows APIs or external tools
	// This is a placeholder for future implementation

	health.ResponseTime = time.Since(start)
	return health
}

// checkExternalAPIHealth performs external API health checks
func (h *HealthService) checkExternalAPIHealth(ctx context.Context) *ComponentHealth {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "external_apis",
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Test OpenRouter API
	openRouterHealthy := h.testAPIEndpoint(ctxWithTimeout, "https://openrouter.ai/api/v1/models", "openrouter")
	health.Details["openrouter"] = openRouterHealthy

	// Test OpenAI API
	openAIHealthy := h.testAPIEndpoint(ctxWithTimeout, "https://api.openai.com/v1/models", "openai")
	health.Details["openai"] = openAIHealthy

	// Determine overall API health
	if openRouterHealthy && openAIHealthy {
		health.Status = HealthStatusHealthy
		health.Message = "All external APIs are accessible"
	} else if openRouterHealthy || openAIHealthy {
		health.Status = HealthStatusDegraded
		health.Message = "Some external APIs are not accessible"
	} else {
		health.Status = HealthStatusUnhealthy
		health.Message = "External APIs are not accessible"
	}

	health.ResponseTime = time.Since(start)
	return health
}

// testAPIEndpoint tests if an API endpoint is accessible
func (h *HealthService) testAPIEndpoint(ctx context.Context, url, name string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"api":   name,
			"error": err.Error(),
		}).Debug("Failed to create API health check request")
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"api":   name,
			"error": err.Error(),
		}).Debug("API health check failed")
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx and 401 (unauthorized) as healthy since the API is responding
	return resp.StatusCode >= 200 && resp.StatusCode < 300 || resp.StatusCode == 401
}

// calculateOverallHealth determines the overall system health based on component health
func (h *HealthService) calculateOverallHealth(components map[string]*ComponentHealth) HealthStatus {
	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, component := range components {
		switch component.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		}
	}

	// If any critical component (database) is unhealthy, system is unhealthy
	if db, exists := components["database"]; exists && db.Status == HealthStatusUnhealthy {
		return HealthStatusUnhealthy
	}

	// If more than half of components are unhealthy, system is unhealthy
	if unhealthyCount > len(components)/2 {
		return HealthStatusUnhealthy
	}

	// If any component is degraded or unhealthy, system is degraded
	if degradedCount > 0 || unhealthyCount > 0 {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// GetComponentHealth returns the health of a specific component
func (h *HealthService) GetComponentHealth(ctx context.Context, componentName string) *ComponentHealth {
	systemHealth := h.GetSystemHealth(ctx)
	if component, exists := systemHealth.Components[componentName]; exists {
		return component
	}
	return &ComponentHealth{
		Name:        componentName,
		Status:      HealthStatusUnhealthy,
		Message:     "Component not found",
		LastChecked: time.Now(),
	}
}

// IsHealthy returns true if the system is healthy
func (h *HealthService) IsHealthy(ctx context.Context) bool {
	return h.GetSystemHealth(ctx).Status == HealthStatusHealthy
}

// ClearCache clears the health check cache, forcing a fresh check on next request
func (h *HealthService) ClearCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cachedHealth = nil
	h.lastCheck = time.Time{}
}
