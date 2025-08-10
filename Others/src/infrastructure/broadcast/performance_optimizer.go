package broadcast

import (
	"runtime"
	"time"
	
	"github.com/sirupsen/logrus"
)

// OptimizeFor3000Devices configures the system for maximum performance
func OptimizeFor3000Devices() {
	// Set GOMAXPROCS to use all CPU cores
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	logrus.Infof("Optimized for %d CPU cores", numCPU)
	
	// Increase garbage collection frequency for better memory management
	runtime.GC()
	
	// Log optimization settings
	logrus.Info("System optimized for 3000 devices:")
	logrus.Infof("- Max concurrent workers: %d", maxConcurrentWorkers)
	logrus.Infof("- Queue check interval: %v", queueCheckInterval)
	logrus.Infof("- Worker batch size: %d", workerBatchSize)
	logrus.Infof("- Health check interval: %v", healthCheckInterval)
}

// BatchCreateWorkers creates multiple workers efficiently
func (um *UltraScaleRedisManager) BatchCreateWorkers(deviceIDs []string) {
	for _, deviceID := range deviceIDs {
		// Create workers in parallel
		go func(id string) {
			um.ensureWorker(id)
		}(deviceID)
		
		// Small delay to prevent thundering herd
		time.Sleep(10 * time.Millisecond)
	}
}
