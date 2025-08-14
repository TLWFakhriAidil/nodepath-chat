# High-Performance Broadcast System

A highly optimized broadcast messaging system designed for WhatsApp automation with support for campaigns, sequences, and bulk messaging. Built for high throughput and reliability.

## Features

### Core Capabilities
- **High-Performance Processing**: Supports 3000+ devices with optimized worker pools
- **Campaign Management**: Queue and manage campaign messages with duplicate prevention
- **Sequence Support**: Handle sequence messages with step delays and ordering
- **Bulk Operations**: Efficient bulk message queuing with transaction support
- **Anti-Spam Protection**: Configurable delays between messages (5-15 seconds default)
- **Health Monitoring**: Real-time metrics and health status tracking
- **Queue Management**: Automatic cleanup of stuck and failed messages
- **Device Management**: Individual device control (pause/resume/restart)

### Performance Optimizations
- **Worker Pool Architecture**: Device-specific workers with large message buffers (1000 messages)
- **Batch Processing**: Process messages in batches for better database performance
- **Connection Pooling**: Optimized database connections with prepared statements
- **Memory Management**: Controlled memory usage with processing time limits
- **Concurrent Processing**: Parallel message processing across multiple devices

## Architecture

### Components

1. **BroadcastService** (`internal/services/broadcast_service.go`)
   - Main orchestration service
   - High-level API for message queuing
   - System management and monitoring

2. **BroadcastManager** (`internal/broadcast/manager.go`)
   - Manages worker pools (up to 500 workers)
   - Handles device-specific message distribution
   - Worker lifecycle management

3. **DeviceWorker** (`internal/broadcast/device_worker.go`)
   - Individual device message processing
   - Anti-spam delay implementation
   - Message status tracking

4. **BroadcastRepository** (`internal/repository/broadcast_repository.go`)
   - Database operations with duplicate prevention
   - Bulk operations support
   - Message status management

5. **QueueCleaner** (`internal/services/queue_cleaner.go`)
   - Automatic cleanup of stuck messages
   - Timeout handling (12-hour limit)
   - Database maintenance

6. **MetricsService** (`internal/services/metrics.go`)
   - Performance monitoring
   - Real-time statistics
   - Health status reporting

## Database Schema

```sql
CREATE TABLE broadcast_messages (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    device_id VARCHAR(36) NOT NULL,
    campaign_id INT NULL,
    sequence_id VARCHAR(36) NULL,
    sequence_stepid VARCHAR(36) NULL,
    recipient_phone VARCHAR(20) NOT NULL,
    recipient_name VARCHAR(255) NULL,
    message_type ENUM('text', 'image', 'video', 'audio', 'document') DEFAULT 'text',
    content TEXT NOT NULL,
    media_url VARCHAR(500) NULL,
    status ENUM('pending', 'queued', 'processing', 'sent', 'failed', 'cancelled') DEFAULT 'pending',
    scheduled_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    group_id VARCHAR(36) NULL,
    group_order INT DEFAULT 0,
    error_message TEXT NULL,
    
    INDEX idx_device_status_scheduled (device_id, status, scheduled_at),
    INDEX idx_campaign_status (campaign_id, status),
    INDEX idx_sequence_status (sequence_id, status),
    INDEX idx_status_created (status, created_at),
    INDEX idx_user_device (user_id, device_id)
);
```

## API Endpoints

### Message Queuing

#### Queue Campaign Message
```http
POST /api/broadcast/campaign/queue
Content-Type: application/json

{
    "user_id": "user123",
    "device_id": "device456",
    "campaign_id": "campaign789",
    "recipient_phone": "+1234567890",
    "message_type": "text",
    "content": "Hello from campaign!",
    "min_delay": 5,
    "max_delay": 15
}
```

#### Queue Sequence Message
```http
POST /api/broadcast/sequence/queue
Content-Type: application/json

{
    "user_id": "user123",
    "device_id": "device456",
    "sequence_id": "seq123",
    "sequence_step_id": "step456",
    "recipient_phone": "+1234567890",
    "message_type": "text",
    "content": "Step 1 of sequence",
    "step_delay": 300
}
```

#### Queue Bulk Messages
```http
POST /api/broadcast/bulk/queue
Content-Type: application/json

{
    "messages": [
        {
            "user_id": "user123",
            "device_id": "device456",
            "campaign_id": "campaign789",
            "recipient_phone": "+1234567890",
            "message_type": "text",
            "content": "Bulk message 1"
        }
    ]
}
```

### Monitoring

#### Get System Metrics
```http
GET /api/broadcast/metrics
```

Response:
```json
{
    "success": true,
    "metrics": {
        "total_processed": 15420,
        "total_failed": 23,
        "total_queued": 1250,
        "messages_per_second": 12.5,
        "average_processing_time": "2.3s",
        "uptime": "2h30m",
        "active_devices": 45,
        "queue_stats": {
            "pending": 1250,
            "processing": 15,
            "sent": 15420,
            "failed": 23
        }
    }
}
```

#### Get Queue Statistics
```http
GET /api/broadcast/queue/stats
```

#### Get Health Status
```http
GET /api/broadcast/health
```

### Device Management

#### Pause Device
```http
POST /api/broadcast/device/{deviceId}/pause
```

#### Resume Device
```http
POST /api/broadcast/device/{deviceId}/resume
```

#### Get Pending Message Count
```http
GET /api/broadcast/device/{deviceId}/pending-count
```

## Usage Examples

### Initialize the System

```go
package main

import (
    "log"
    "nodepath-chat/internal/services"
)

func main() {
    // Create broadcast service
    broadcastService := services.NewBroadcastService()
    
    // Start the service
    err := broadcastService.Start()
    if err != nil {
        log.Fatalf("Failed to start broadcast service: %v", err)
    }
    
    defer broadcastService.Stop()
    
    // Service is now ready to handle messages
    log.Println("Broadcast service started successfully")
}
```

### Queue Campaign Messages

```go
// Queue a single campaign message
messageID, err := broadcastService.QueueCampaignMessage(
    "user123",           // userID
    "device456",         // deviceID
    "campaign789",       // campaignID
    "+1234567890",       // recipientPhone
    "text",              // messageType
    "Hello from campaign!", // content
    "",                  // mediaURL
    5,                   // minDelay (seconds)
    15,                  // maxDelay (seconds)
)
if err != nil {
    log.Printf("Failed to queue message: %v", err)
} else {
    log.Printf("Message queued with ID: %s", messageID)
}
```

### Queue Sequence Messages

```go
// Queue a sequence message with step delay
messageID, err := broadcastService.QueueSequenceMessage(
    "user123",           // userID
    "device456",         // deviceID
    "seq123",            // sequenceID
    "step456",           // sequenceStepID
    "+1234567890",       // recipientPhone
    "text",              // messageType
    "Step 1 of sequence", // content
    "",                  // mediaURL
    5,                   // minDelay (seconds)
    15,                  // maxDelay (seconds)
    300*time.Second,     // stepDelay
)
```

### Bulk Message Queuing

```go
// Prepare bulk messages
messages := []models.BroadcastMessage{
    {
        UserID:         "user123",
        DeviceID:       "device456",
        CampaignID:     &campaignID,
        RecipientPhone: "+1234567890",
        Type:           "text",
        Content:        "Bulk message 1",
        MinDelay:       5,
        MaxDelay:       15,
    },
    // ... more messages
}

// Queue all messages in a single transaction
messageIDs, err := broadcastService.QueueBulkMessages(messages)
if err != nil {
    log.Printf("Failed to queue bulk messages: %v", err)
} else {
    log.Printf("Queued %d messages", len(messageIDs))
}
```

### Monitor System Performance

```go
// Get current metrics
metrics := broadcastService.GetMetrics()
log.Printf("Messages per second: %.2f", metrics.MessagesPerSecond)
log.Printf("Active devices: %d", metrics.ActiveDevices)
log.Printf("Total processed: %d", metrics.TotalProcessed)

// Get health status
healthStatus := broadcastService.GetHealthStatus()
log.Printf("System health: %s", healthStatus)

// Get queue statistics
stats, err := broadcastService.GetQueueStats()
if err == nil {
    log.Printf("Pending messages: %d", stats["pending"])
    log.Printf("Failed messages: %d", stats["failed"])
}
```

## Performance Tuning

### Worker Configuration
- **Max Workers**: Default 500, can be adjusted based on system resources
- **Queue Buffer**: 1000 messages per worker, increase for higher throughput
- **Batch Size**: Default 50 messages per batch, optimize based on database performance

### Database Optimization
- Use connection pooling with appropriate limits
- Ensure proper indexing on frequently queried columns
- Consider partitioning for very large message volumes
- Regular cleanup of old messages to maintain performance

### Memory Management
- Monitor processing time storage (default 1000 entries)
- Adjust cleanup intervals based on message volume
- Use appropriate timeouts for stuck message detection

## Monitoring and Alerting

### Key Metrics to Monitor
- **Messages per second**: Throughput indicator
- **Failure rate**: Quality indicator (should be < 5%)
- **Queue depth**: Backlog indicator
- **Active devices**: Capacity indicator
- **Processing time**: Performance indicator

### Health Checks
- System returns "healthy", "warning", or "critical"
- Based on failure rates, throughput, and active devices
- Integrate with monitoring systems for alerting

## Troubleshooting

### Common Issues

1. **High Failure Rate**
   - Check WhatsApp API connectivity
   - Verify device authentication
   - Review error messages in database

2. **Low Throughput**
   - Increase worker count
   - Optimize database queries
   - Check system resources

3. **Stuck Messages**
   - Queue cleaner runs every 5 minutes
   - Manual cleanup available via API
   - Check worker health status

### Debugging

```go
// Get worker status
workerStatus := broadcastService.GetWorkerStatus()
for _, status := range workerStatus {
    log.Printf("Device %s: %s (Queue: %d, Processed: %d)", 
        status.DeviceID, status.Status, status.QueueSize, status.ProcessedCount)
}

// Force cleanup if needed
broadcastService.ForceCleanup()

// Restart problematic device
err := broadcastService.RestartDevice("device456")
if err != nil {
    log.Printf("Failed to restart device: %v", err)
}
```

## Security Considerations

- Validate all input parameters
- Implement rate limiting on API endpoints
- Use secure database connections
- Log security events and failures
- Implement proper authentication and authorization

## Scalability

The system is designed to scale horizontally:
- Database can be sharded by device_id
- Multiple application instances can run concurrently
- Redis can be used for distributed coordination
- Load balancing across multiple servers

For very high volumes (10,000+ devices), consider:
- Microservice architecture
- Message queue systems (RabbitMQ, Apache Kafka)
- Distributed caching
- Database clustering