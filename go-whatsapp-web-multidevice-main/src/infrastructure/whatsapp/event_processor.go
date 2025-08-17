package whatsapp

import (
	"runtime"
	"sync"
	"time"
	"github.com/sirupsen/logrus"
)

// EventProcessor handles async event processing with worker pools
type EventProcessor struct {
	workers    int
	eventQueue chan EventTask
	wg         sync.WaitGroup
	started    bool
	mu         sync.Mutex
}

// EventTask represents an event to be processed
type EventTask struct {
	DeviceID string
	Event    interface{}
	Handler  func(interface{})
}

var (
	globalEventProcessor *EventProcessor
	processorOnce        sync.Once
)

// GetEventProcessor returns the global event processor
func GetEventProcessor() *EventProcessor {
	processorOnce.Do(func() {
		// Create processor with workers based on CPU cores
		numWorkers := runtime.NumCPU() * 4 // 4 workers per CPU core
		if numWorkers > 100 {
			numWorkers = 100 // Cap at 100 workers
		}
		
		globalEventProcessor = &EventProcessor{
			workers:    numWorkers,
			eventQueue: make(chan EventTask, 10000), // Large buffer
		}
		
		globalEventProcessor.Start()
		logrus.Infof("Event processor started with %d workers", numWorkers)
	})
	
	return globalEventProcessor
}

// Start initializes the worker pool
func (ep *EventProcessor) Start() {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	
	if ep.started {
		return
	}
	
	// Start worker goroutines
	for i := 0; i < ep.workers; i++ {
		ep.wg.Add(1)
		go ep.worker(i)
	}
	
	ep.started = true
	
	// Start monitoring
	go ep.monitor()
}

// worker processes events from the queue
func (ep *EventProcessor) worker(id int) {
	defer ep.wg.Done()
	
	for task := range ep.eventQueue {
		ep.processEvent(task)
	}
}

// processEvent handles a single event
func (ep *EventProcessor) processEvent(task EventTask) {
	// Recover from any panics in event processing
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Panic in event processing for device %s: %v", task.DeviceID, r)
		}
	}()
	
	// Process the event
	if task.Handler != nil {
		task.Handler(task.Event)
	}
}

// QueueEvent adds an event to the processing queue
func (ep *EventProcessor) QueueEvent(deviceID string, event interface{}, handler func(interface{})) {
	select {
	case ep.eventQueue <- EventTask{
		DeviceID: deviceID,
		Event:    event,
		Handler:  handler,
	}:
		// Successfully queued
	default:
		// Queue full, log and drop
		logrus.Warnf("Event queue full, dropping event for device %s", deviceID)
	}
}

// monitor provides statistics about the event processor
func (ep *EventProcessor) monitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		queueSize := len(ep.eventQueue)
		if queueSize > 5000 {
			logrus.Warnf("Event queue size high: %d events pending", queueSize)
		} else if queueSize > 0 {
			logrus.Debugf("Event queue size: %d", queueSize)
		}
	}
}

// GetQueueSize returns current queue size
func (ep *EventProcessor) GetQueueSize() int {
	return len(ep.eventQueue)
}

// Shutdown gracefully stops the event processor
func (ep *EventProcessor) Shutdown() {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	
	if !ep.started {
		return
	}
	
	// Close the queue
	close(ep.eventQueue)
	
	// Wait for workers to finish
	ep.wg.Wait()
	
	ep.started = false
	logrus.Info("Event processor shut down")
}
