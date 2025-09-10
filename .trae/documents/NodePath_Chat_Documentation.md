# NodePath Chat System Documentation

## Table of Contents

1. [Project Overview](#project-overview)
2. [System Architecture](#system-architecture)
3. [Technology Stack](#technology-stack)
4. [API Documentation](#api-documentation)
5. [Database Schema](#database-schema)
6. [AI System Documentation](#ai-system-documentation)
7. [Flow Builder Documentation](#flow-builder-documentation)
8. [Development Guidelines](#development-guidelines)
9. [Deployment Guide](#deployment-guide)
10. [Troubleshooting Guide](#troubleshooting-guide)

---

## 1. Project Overview

NodePath Chat is a high-performance WhatsApp chatbot system designed to handle 3000+ concurrent devices with real-time AI-powered conversations. The system provides intelligent conversation flows, dynamic AI responses, and comprehensive analytics.

### Key Features
- **Real-time Processing**: Handles multiple concurrent WhatsApp conversations
- **AI Integration**: OpenRouter/OpenAI powered intelligent responses
- **Flow Builder**: Visual flow creation with multiple node types
- **Multi-Provider Support**: Wablas and Whacenter WhatsApp providers
- **Analytics Dashboard**: Comprehensive conversation and lead analytics
- **Device Management**: Individual device configurations and settings

### Target Performance
- **Concurrent Devices**: 3000+ simultaneous connections
- **Response Time**: Sub-second AI response generation
- **Uptime**: 99.9% availability on Railway platform
- **Database**: MySQL 5.7 with optimized queries

---

## 2. System Architecture

```mermaid
graph TD
    A[WhatsApp Users] --> B[WhatsApp Providers]
    B --> C[NodePath Chat System]
    C --> D[AI Services]
    C --> E[MySQL Database]
    C --> F[Redis Cache]
    
    subgraph "WhatsApp Providers"
        B1[Wablas API]
        B2[Whacenter API]
    end
    
    subgraph "NodePath Chat System"
        C1[Go Backend Server]
        C2[React Frontend]
        C3[Webhook Handlers]
        C4[Flow Engine]
        C5[Analytics Engine]
    end
    
    subgraph "AI Services"
        D1[OpenRouter API]
        D2[OpenAI API]
    end
    
    subgraph "Data Layer"
        E1[Conversation Data]
        E2[Flow Definitions]
        E3[Device Settings]
        E4[Analytics Data]
    end
```

### Component Responsibilities

- **Go Backend**: Core processing, webhook handling, AI integration
- **React Frontend**: Flow builder, analytics dashboard, device management
- **MySQL Database**: Persistent data storage with _nodepath suffix tables
- **Redis Cache**: Session management and temporary data
- **AI Services**: Intelligent response generation

---

## 3. Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin HTTP framework
- **Database**: MySQL 5.7
- **Cache**: Redis
- **Deployment**: Railway platform
- **Port**: 8080 (both local and production)

### Frontend
- **Framework**: React 18
- **Styling**: Tailwind CSS
- **Build Tool**: Vite
- **UI Components**: Custom component library

### External Services
- **AI Provider**: OpenRouter (primary), OpenAI (fallback)
- **WhatsApp Providers**: Wablas, Whacenter
- **Database Host**: Railway MySQL

### Environment Configuration
```bash
# Database Connection
MYSQL_URI=mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway

# Server Configuration
PORT=8080
ENVIRONMENT=production

# AI Configuration
OPENROUTER_API_URL=https://openrouter.ai/api/v1/chat/completions
OPENAI_API_URL=https://api.openai.com/v1/chat/completions
```

---

## 4. API Documentation

### WhatsApp Webhook Endpoints

#### Receive Message
```http
POST /webhook/whatsapp
Content-Type: application/json

{
  "id_device": "FakhriAidilTLW-001",
  "phone_number": "601137508067",
  "message": "Hello",
  "message_type": "text",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Send Message Response
```http
POST /api/whatsapp/send
Content-Type: application/json
Authorization: Bearer {device_api_key}

{
  "id_device": "FakhriAidilTLW-001",
  "phone_number": "601137508067",
  "message": "AI generated response",
  "message_type": "text"
}
```

### AI Processing Endpoints

#### Generate AI Response
```http
POST /api/ai/generate
Content-Type: application/json

{
  "id_device": "FakhriAidilTLW-001",
  "flow_id": "flow_ai_1756016272",
  "current_text": "User message",
  "last_text": "Previous bot response",
  "stage": "Problem Identification"
}
```

### Flow Management Endpoints

#### Get Flow Definition
```http
GET /api/flows/{flow_id}
Authorization: Bearer {device_api_key}
```

#### Update Flow
```http
PUT /api/flows/{flow_id}
Content-Type: application/json

{
  "flow_data": {
    "nodes": [...],
    "connections": [...]
  }
}
```

---

## 5. Database Schema

### Core Tables (All end with _nodepath)

#### device_settings_nodepath
```sql
CREATE TABLE device_settings_nodepath (
    id_device VARCHAR(50) PRIMARY KEY,
    device_name VARCHAR(100) NOT NULL,
    api_key_option VARCHAR(100),
    api_provider VARCHAR(20) DEFAULT 'openrouter',
    whatsapp_provider VARCHAR(20) DEFAULT 'wablas',
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### chatbot_flows_nodepath
```sql
CREATE TABLE chatbot_flows_nodepath (
    flow_id VARCHAR(50) PRIMARY KEY,
    id_device VARCHAR(50),
    flow_name VARCHAR(100),
    flow_data JSON,
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_device) REFERENCES device_settings_nodepath(id_device)
);
```

#### ai_whatsapp_nodepath
```sql
CREATE TABLE ai_whatsapp_nodepath (
    id_prospect INT AUTO_INCREMENT PRIMARY KEY,
    id_device VARCHAR(50),
    prospect_num VARCHAR(20),
    prospect_nama VARCHAR(100),
    conv_current TEXT,
    conv_last TEXT,
    stage VARCHAR(100),
    flow_id VARCHAR(50),
    human_status TINYINT DEFAULT 0,
    balas TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_device) REFERENCES device_settings_nodepath(id_device)
);
```

#### conversation_logs_nodepath
```sql
CREATE TABLE conversation_logs_nodepath (
    log_id INT AUTO_INCREMENT PRIMARY KEY,
    id_device VARCHAR(50),
    phone_number VARCHAR(20),
    message_type ENUM('incoming', 'outgoing'),
    content TEXT,
    ai_response_time DECIMAL(5,3),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_device) REFERENCES device_settings_nodepath(id_device)
);
```

### Relationships
- `device_settings_nodepath` is the central table
- All other tables reference `id_device` as foreign key
- `chatbot_flows_nodepath` stores JSON flow definitions
- `ai_whatsapp_nodepath` maintains conversation state
- `conversation_logs_nodepath` tracks all message exchanges

---

## 6. AI System Documentation

### AI Processing Flow

```mermaid
sequenceDiagram
    participant User as WhatsApp User
    participant WH as Webhook Handler
    participant AI as AI Service
    participant OR as OpenRouter/OpenAI
    participant DB as Database
    
    User->>WH: Send Message
    WH->>DB: Get Device Settings
    WH->>DB: Get Flow Definition
    WH->>AI: Process AI Node
    AI->>DB: Get Conversation History
    AI->>OR: Generate Response
    OR->>AI: AI Response
    AI->>DB: Update Conversation
    AI->>WH: Processed Response
    WH->>User: Send Reply
```

### AI Payload Structure

The system uses a standardized payload for all AI requests:

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "AI prompt with instructions"
    },
    {
      "role": "assistant",
      "content": "Previous bot response"
    },
    {
      "role": "user",
      "content": "Current user message"
    }
  ],
  "temperature": 0.67,
  "top_p": 1,
  "repetition_penalty": 1
}
```

### AI Response Format

All AI responses must follow this JSON structure:

```json
{
  "Stage": "Problem Identification",
  "Response": [
    {
      "type": "text",
      "Jenis": "onemessage",
      "content": "First response message"
    },
    {
      "type": "image",
      "content": "https://example.com/image.jpg"
    },
    {
      "type": "text",
      "Jenis": "onemessage",
      "content": "Second response message"
    }
  ]
}
```

### Stage Management

- **Stage Detection**: Automatic stage progression based on user input
- **Stage Override**: Manual stage setting via "stage: [stage_name]" command
- **Default Stage**: Falls back to first stage if undefined
- **Stage Persistence**: Stored in `ai_whatsapp_nodepath.stage`

### "onemessage" Logic

When AI responses include `"Jenis": "onemessage"` in text parts:
1. System collects all consecutive text parts with this flag
2. Combines them into a single message with 2-second delays
3. Logs as "BOT_COMBINED" in conversation logs
4. Sends as one cohesive response to user

### API Provider Selection

```go
// API URL selection logic
func getAPIURL(deviceID string) string {
    if deviceID == "SCHQ-S94" || deviceID == "SCHQ-S12" {
        return "https://api.openai.com/v1/chat/completions"
    }
    return "https://openrouter.ai/api/v1/chat/completions"
}
```

---

## 7. Flow Builder Documentation

### Node Types

#### 1. AI Prompt Node
- **Purpose**: Generate AI-powered responses
- **Configuration**: System prompt, model selection, temperature
- **Processing**: Uses single `processAIPromptNode` function
- **Output**: JSON structured response with stage and content

#### 2. Message Node
- **Purpose**: Send predefined text messages
- **Configuration**: Message content with variable support
- **Variables**: `{{phone_number}}`, `{{name}}`, `{{stage}}`
- **Processing**: Direct message sending with variable replacement

#### 3. Image Node
- **Purpose**: Send image messages
- **Configuration**: Image URL or base64 data
- **Support**: JPG, PNG, GIF formats
- **Processing**: Media upload and message sending

#### 4. Delay Node
- **Purpose**: Add timing delays between messages
- **Configuration**: Delay duration in seconds
- **Range**: 1-60 seconds
- **Processing**: Non-blocking delay implementation

#### 5. Condition Node
- **Purpose**: Branch flow based on user input
- **Configuration**: Condition rules and target nodes
- **Logic**: Text matching, keyword detection, regex patterns
- **Processing**: Dynamic flow routing

### Flow Execution Logic

```mermaid
flowchart TD
    A[Start Node] --> B{Node Type?}
    B -->|AI Prompt| C[Process AI Node]
    B -->|Message| D[Send Message]
    B -->|Image| E[Send Image]
    B -->|Delay| F[Wait Duration]
    B -->|Condition| G[Evaluate Condition]
    
    C --> H[Update Stage]
    D --> I[Replace Variables]
    E --> J[Upload Media]
    F --> K[Continue Flow]
    G --> L{Condition Met?}
    
    H --> M[Next Node]
    I --> M
    J --> M
    K --> M
    L -->|Yes| N[Target Node]
    L -->|No| O[Default Node]
    
    M --> P[End or Continue]
    N --> P
    O --> P
```

### Variable System

Supported variables in message and AI prompt nodes:

- `{{phone_number}}`: User's WhatsApp number
- `{{name}}`: User's display name
- `{{stage}}`: Current conversation stage
- `{{device_id}}`: Device identifier
- `{{flow_id}}`: Current flow identifier
- `{{timestamp}}`: Current timestamp

### Flow Configuration Example

```json
{
  "flow_id": "flow_ai_1756016272",
  "flow_name": "Customer Support Flow",
  "nodes": [
    {
      "id": "start",
      "type": "ai_prompt",
      "prompt": "You are a helpful customer support agent...",
      "model": "gpt-3.5-turbo",
      "next_node": "condition_1"
    },
    {
      "id": "condition_1",
      "type": "condition",
      "rules": [
        {
          "condition": "contains",
          "value": "problem",
          "target_node": "support_flow"
        }
      ],
      "default_node": "general_info"
    }
  ]
}
```

---

## 8. Development Guidelines

### Code Structure

```
nodepath-chat-1/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models and structs
│   ├── repository/      # Database access layer
│   ├── services/        # Business logic services
│   ├── utils/           # Utility functions
│   └── whatsapp/        # WhatsApp integration
├── src/                 # React frontend source
├── migrations/          # Database migrations
└── docs/               # Documentation
```

### Coding Standards

#### Go Backend
- **Function Comments**: All public functions must have comments
- **Error Handling**: Always handle errors explicitly
- **Logging**: Use structured logging with appropriate levels
- **Testing**: Unit tests for all business logic
- **Naming**: Use camelCase for variables, PascalCase for types

```go
// ProcessAIPromptNode handles AI prompt node execution
// Returns processed response and any execution error
func ProcessAIPromptNode(ctx context.Context, node *models.AIPromptNode) (*models.AIResponse, error) {
    // Implementation with proper error handling
    if err := validateNode(node); err != nil {
        return nil, fmt.Errorf("node validation failed: %w", err)
    }
    // ... rest of implementation
}
```

#### React Frontend
- **Component Structure**: Functional components with hooks
- **TypeScript**: Strict type checking enabled
- **Styling**: Tailwind CSS utility classes
- **State Management**: React Context for global state

### Testing Procedures

#### Backend Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/services -run TestAIService
```

#### Frontend Testing
```bash
# Run React tests
npm test

# Run with coverage
npm run test:coverage
```

#### Integration Testing
```bash
# Test with specific device
id_device="FakhriAidilTLW-001"
flow_id="flow_ai_1756016272"
phone="601137508067"

# Test webhook endpoint
curl -X POST http://localhost:8080/webhook/whatsapp \
  -H "Content-Type: application/json" \
  -d '{
    "id_device": "'$id_device'",
    "phone_number": "'$phone'",
    "message": "Hello"
  }'
```

### Performance Optimization

#### Database Optimization
- **Indexing**: Create indexes on frequently queried columns
- **Connection Pooling**: Use connection pools for database access
- **Query Optimization**: Use EXPLAIN to analyze query performance

```sql
-- Example indexes for performance
CREATE INDEX idx_ai_whatsapp_device_phone ON ai_whatsapp_nodepath(id_device, prospect_num);
CREATE INDEX idx_conversation_logs_device_time ON conversation_logs_nodepath(id_device, created_at);
```

#### Caching Strategy
- **Redis**: Cache frequently accessed data
- **Memory Cache**: In-memory caching for static data
- **TTL**: Appropriate time-to-live for cached data

---

## 9. Deployment Guide

### Railway Deployment

#### Environment Setup
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Link to project
railway link
```

#### Environment Variables
```bash
# Set required environment variables
railway variables set MYSQL_URI="mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
railway variables set PORT="8080"
railway variables set ENVIRONMENT="production"
```

#### Deployment Process
```bash
# Deploy to Railway
railway up

# Check deployment status
railway status

# View logs
railway logs
```

### Local Development Setup

#### Prerequisites
- Go 1.21+
- Node.js 18+
- MySQL 5.7+
- Redis (optional)

#### Setup Steps
```bash
# Clone repository
git clone <repository-url>
cd nodepath-chat-1

# Install Go dependencies
go mod download

# Install Node dependencies
npm install

# Set environment variables
cp .env.example .env
# Edit .env with your configuration

# Run database migrations
go run cmd/migrate/main.go

# Start backend server
go run cmd/server/main.go

# Start frontend (in another terminal)
npm run dev
```

### Build Process

#### Backend Build
```bash
# Build for production
go build -o bin/server cmd/server/main.go

# Build without CGO (for Railway)
CGO_ENABLED=0 go build -o bin/server cmd/server/main.go
```

#### Frontend Build
```bash
# Build React app
npm run build

# Preview build
npm run preview
```

---

## 10. Troubleshooting Guide

### Common Issues

#### 1. Database Connection Issues

**Symptoms**: Connection timeouts, authentication errors

**Solutions**:
```bash
# Check database connectivity
mysql -h 157.245.206.124 -u admin_aqil -p admin_railway

# Verify environment variables
echo $MYSQL_URI

# Test connection in Go
go run debug/test_db_connection.go
```

#### 2. AI API Failures

**Symptoms**: Empty responses, API errors, timeout issues

**Solutions**:
```bash
# Check API key validity
curl -H "Authorization: Bearer $API_KEY" \
     https://openrouter.ai/api/v1/models

# Verify payload structure
go run debug/test_ai_payload.go

# Check rate limits
grep "rate limit" /var/log/app.log
```

#### 3. WhatsApp Webhook Issues

**Symptoms**: Messages not received, webhook timeouts

**Solutions**:
```bash
# Test webhook endpoint
curl -X POST http://localhost:8080/webhook/whatsapp \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'

# Check webhook logs
grep "webhook" /var/log/app.log

# Verify provider configuration
go run debug/test_whatsapp_provider.go
```

#### 4. Performance Issues

**Symptoms**: Slow response times, high memory usage

**Solutions**:
```bash
# Monitor system resources
top -p $(pgrep server)

# Check database performance
SHOW PROCESSLIST;
SHOW STATUS LIKE 'Slow_queries';

# Profile Go application
go tool pprof http://localhost:8080/debug/pprof/profile
```

### Error Handling Patterns

#### Graceful Degradation
```go
// Example: Fallback when AI service fails
func (s *AIService) GenerateResponse(ctx context.Context, req *AIRequest) (*AIResponse, error) {
    response, err := s.callPrimaryAPI(ctx, req)
    if err != nil {
        log.Warn("Primary AI API failed, trying fallback", "error", err)
        response, err = s.callFallbackAPI(ctx, req)
        if err != nil {
            return s.getDefaultResponse(req), nil // Graceful fallback
        }
    }
    return response, nil
}
```

#### Circuit Breaker Pattern
```go
// Prevent cascading failures
type CircuitBreaker struct {
    failureCount int
    lastFailTime time.Time
    state        string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == "open" && time.Since(cb.lastFailTime) < 30*time.Second {
        return errors.New("circuit breaker open")
    }
    // ... implementation
}
```

### Monitoring and Alerts

#### Key Metrics to Monitor
- **Response Time**: API response times < 2 seconds
- **Error Rate**: < 1% error rate for critical endpoints
- **Database Connections**: Monitor connection pool usage
- **Memory Usage**: Keep below 80% of available memory
- **Disk Space**: Monitor log file growth

#### Log Analysis
```bash
# Find errors in logs
grep -i error /var/log/app.log | tail -20

# Monitor response times
grep "response_time" /var/log/app.log | awk '{print $NF}' | sort -n

# Check AI processing times
grep "AI_PROCESSING_TIME" /var/log/app.log
```

### Recovery Procedures

#### Database Recovery
```sql
-- Check table status
SHOW TABLE STATUS LIKE '%_nodepath';

-- Repair corrupted tables
REPAIR TABLE ai_whatsapp_nodepath;

-- Optimize tables
OPTIMIZE TABLE conversation_logs_nodepath;
```

#### Service Recovery
```bash
# Restart application
sudo systemctl restart nodepath-chat

# Clear cache
redis-cli FLUSHALL

# Reset database connections
killall -USR1 server
```

---

## Support and Maintenance

### Regular Maintenance Tasks

1. **Database Cleanup**: Remove old conversation logs (>30 days)
2. **Log Rotation**: Rotate application logs weekly
3. **Cache Cleanup**: Clear expired cache entries
4. **Performance Review**: Monthly performance analysis
5. **Security Updates**: Keep dependencies updated

### Contact Information

- **Development Team**: [team@nodepath.com]
- **System Administrator**: [admin@nodepath.com]
- **Emergency Contact**: [emergency@nodepath.com]

### Version History

- **v1.0.0**: Initial release with basic AI chat functionality
- **v1.1.0**: Added flow builder and analytics
- **v1.2.0**: Performance optimizations for 3000+ devices
- **v1.3.0**: Enhanced AI processing with onemessage support

---

*Last Updated: January 2024*
*Documentation Version: 1.3.0*