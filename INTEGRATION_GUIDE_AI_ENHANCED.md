# Integration Guide - Enhanced AI Response System

## Overview
This guide shows how to integrate the enhanced AI response system with onemessage combining, stage detection, and improved response handling into your existing NodePath Chat system.

## Files Created

1. **ai_response_parser_enhanced.go** - Enhanced parser with multiple format support
2. **ai_onemessage_processor.go** - Onemessage combining logic
3. **ai_prompt_builder.go** - Enhanced prompt building with instructions
4. **ai_conversation_enhanced.go** - Complete enhanced conversation processing

## Integration Steps

### Step 1: Update AIWhatsappResponseItem Model
In `internal/models/models.go`, update the response item structure:

```go
// AIWhatsappResponseItem represents individual response items
type AIWhatsappResponseItem struct {
    Type    string `json:"type"`
    Jenis   string `json:"Jenis,omitempty"`  // Add this field for onemessage support
    Content string `json:"content"`
}
```

### Step 2: Add Balas Field to AIWhatsapp Model
In `internal/models/models.go`, add the Balas field for time tracking:

```go
type AIWhatsapp struct {
    // ... existing fields ...
    Balas       sql.NullString `json:"balas" db:"balas"`  // Add this for timestamp tracking
    // ... other fields ...
}
```

### Step 3: Update AIWhatsappService Interface
In `internal/services/ai_whatsapp_service.go`, add new method signatures:

```go
type AIWhatsappService interface {
    // ... existing methods ...
    
    // Enhanced methods
    ParseAIResponseEnhanced(responseText string, prospectNum string, idDevice string) (*AIWhatsappResponse, error)
    ProcessAIResponseWithOnemessage(response *AIWhatsappResponse, prospectNum string, deviceID string, sendMessageFunc func(phone, message string) error, sendMediaFunc func(phone, mediaURL string) error) error
    ProcessAIConversationEnhanced(prospectNum string, idDevice string, currentText string, stage string) (*AIWhatsappResponse, error)
    BuildEnhancedAIPrompt(behavePrompt string, closingPrompt string, currentStage string) string
    ExtractStageFromUserInput(userInput string) (string, bool)
    CheckTimeThrottle(lastResponseTime time.Time, thresholdSeconds int) bool
}
```

### Step 4: Integrate Enhanced Functions
Copy the created files into your service layer:

```bash
# Copy enhanced functions to services directory
cp ai_response_parser_enhanced.go internal/services/
cp ai_onemessage_processor.go internal/services/
cp ai_prompt_builder.go internal/services/
cp ai_conversation_enhanced.go internal/services/
```

### Step 5: Update Webhook Handler
In `internal/handlers/device_settings_handlers.go`, update the processAIConversation function:

```go
func (h *DeviceSettingsHandlers) processAIConversation(phoneNumber, content, deviceID string) {
    // Extract stage from user input if present
    stage := ""
    if detectedStage, found := h.aiWhatsappService.ExtractStageFromUserInput(content); found {
        stage = detectedStage
        logrus.WithFields(logrus.Fields{
            "stage": stage,
            "phone": phoneNumber,
        }).Info("🎯 Detected stage from user input")
    }
    
    // Process enhanced AI conversation
    aiResponse, err := h.aiWhatsappService.ProcessAIConversationEnhanced(
        phoneNumber,
        deviceID,
        content,
        stage,
    )
    
    if err != nil {
        logrus.WithError(err).Error("Failed to process AI conversation")
        return
    }
    
    // Process response with onemessage combining
    err = h.aiWhatsappService.ProcessAIResponseWithOnemessage(
        aiResponse,
        phoneNumber,
        deviceID,
        func(phone, message string) error {
            // Use your existing send message function
            return h.whatsappService.SendMessageFromDevice(deviceID, phone, message)
        },
        func(phone, mediaURL string) error {
            // Use your existing send media function
            return h.whatsappService.SendMediaMessage(deviceID, phone, mediaURL)
        },
    )
    
    if err != nil {
        logrus.WithError(err).Error("Failed to process AI response")
    }
}
```

### Step 6: Database Migration
Create a migration to add the balas field if it doesn't exist:

```sql
-- Add balas field for timestamp tracking
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS balas VARCHAR(255) DEFAULT NULL;
```

### Step 7: Testing

Test the enhanced system with various inputs:

1. **Test Onemessage Combining**:
   - Send: "I want this section in add response format [onemessage]"
   - Verify consecutive text parts are combined

2. **Test Stage Detection**:
   - Send: "stage: Problem Identification"
   - Verify stage is extracted and updated

3. **Test Time Throttling**:
   - Send rapid messages
   - Verify 4-second throttling is enforced

4. **Test Multiple Response Formats**:
   - Test JSON responses
   - Test legacy format responses
   - Test plain text fallback

## Key Features Implemented

### 1. ✅ Onemessage Combining
- Detects `Jenis: "onemessage"` in response items
- Combines consecutive onemessage items with newlines
- Logs as BOT_COMBINED in conversation history
- Handles mixed response types correctly

### 2. ✅ Stage Detection
- Extracts stage from user messages using regex
- Pattern: `/\bstage\s*:\s*(.+)/i`
- Automatically updates conversation stage
- Maintains stage throughout conversation

### 3. ✅ Time Throttling
- 4-second minimum between responses
- Prevents rapid consecutive messages
- Uses balas field for timestamp tracking
- Returns throttle error when too soon

### 4. ✅ Enhanced Prompt Structure
- Includes behavior and closing prompts
- Instructions for onemessage handling
- Stage-aware response generation
- Prevention of repetitive responses

### 5. ✅ Multiple Response Format Support
- Standard JSON with Stage and Response
- Legacy format with prefixes
- Plain text fallback
- JSON wrapped in backticks
- Encapsulated JSON in content

### 6. ✅ Conversation Logging
- BOT_COMBINED for combined messages
- BOT for regular text messages
- BOT_IMAGE, BOT_AUDIO, BOT_VIDEO for media
- Maintains conversation history in conv_last

## Performance Considerations

1. **Response Time Logging**: Tracks API response times for monitoring
2. **Async Processing**: Uses goroutines for message processing
3. **Error Handling**: Comprehensive error handling with logging
4. **Fallback Support**: Graceful degradation for various response formats

## Troubleshooting

### Common Issues:

1. **Onemessage not combining**:
   - Check if Jenis field is properly set in response
   - Verify response parsing is working correctly

2. **Stage not detected**:
   - Check regex pattern matches your input format
   - Verify case-insensitive matching is working

3. **Time throttling too strict**:
   - Adjust threshold in CheckTimeThrottle function
   - Default is 4 seconds, can be configured

4. **Response parsing fails**:
   - Check JSON format is valid
   - Try different response formats
   - Check logs for parsing errors

## Conclusion

The enhanced AI response system is now integrated with your NodePath Chat application, providing:
- Advanced response combining with onemessage support
- Automatic stage detection from user input
- Time-based throttling to prevent spam
- Multiple response format support
- Enhanced conversation logging with BOT_COMBINED entries

The system maintains backward compatibility while adding these new features from your PHP implementation.