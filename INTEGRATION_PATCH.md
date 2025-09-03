# ENHANCED AI INTEGRATION PATCH

## How to Use the Enhanced AI Methods

The enhanced AI methods are already available in `ai_whatsapp_enhanced.go`. Here's how to integrate them:

### 1. In Your ProcessAIConversation Flow:

```go
// At the beginning of ProcessAIConversation
func (s *aiWhatsappService) ProcessAIConversation(...) (*AIWhatsappResponse, error) {
    // Add stage detection
    if detectedStage, found := s.ExtractStageFromUserInput(currentText); found {
        stage = detectedStage
        logrus.Info("Detected stage from user input: ", stage)
    }
    
    // After getting aiConv from database
    if aiConv != nil && aiConv.Balas != "" {
        // Check time throttling
        if !s.CheckTimeThrottle(aiConv.Balas, 4) {
            return nil, fmt.Errorf("request throttled")
        }
    }
    
    // ... rest of processing
    
    // After AI response, update balas timestamp
    aiConv.Balas = time.Now().Format("2006-01-02 15:04:05")
    s.aiRepo.UpdateAIWhatsapp(aiConv)
}
```

### 2. In Your Webhook Handler:

```go
// In processAIConversation handler
func (h *Handlers) processAIConversation(...) {
    // Get AI response
    response, err := h.aiWhatsappService.ProcessAIConversation(...)
    
    if response != nil {
        // Use enhanced onemessage processing
        err = h.aiWhatsappService.ProcessAIResponseWithOnemessage(
            response,
            from,
            idDevice,
            func(phone, text string) error {
                // Send text through your provider
                return h.sendTextMessage(phone, text, idDevice, provider)
            },
            func(phone, mediaURL string) error {
                // Send media through your provider
                return h.sendMediaMessage(phone, mediaURL, idDevice, provider)
            },
        )
    }
}
```

### 3. Enhanced Prompt Building:

```go
// When building AI prompt
systemPrompt := s.BuildEnhancedAIPrompt(
    behavePrompt,   // Your behavior prompt
    closingPrompt,  // Your closing prompt
    currentStage    // Current conversation stage
)

// Use this in your AI API call payload
payload := map[string]interface{}{
    "model": model,
    "messages": []map[string]interface{}{
        {"role": "system", "content": systemPrompt},
        {"role": "assistant", "content": lastText},
        {"role": "user", "content": currentText},
    },
    "temperature": 0.67,
    "top_p": 1.0,
}
```

## Available Enhanced Methods:

1. **ExtractStageFromUserInput(userInput string) (string, bool)**
   - Detects stage from messages like "stage: Problem Identification"
   - Returns stage and true if found

2. **CheckTimeThrottle(lastResponseTime string, thresholdSeconds int) bool**
   - Checks if enough time has passed since last response
   - Uses format "2006-01-02 15:04:05"
   - Returns false if throttled

3. **ProcessAIResponseWithOnemessage(...)**
   - Handles onemessage combining
   - Logs as BOT_COMBINED for combined messages
   - Updates balas timestamp automatically

4. **BuildEnhancedAIPrompt(behavePrompt, closingPrompt, currentStage string) string**
   - Creates prompt with onemessage instructions
   - Includes stage management rules
   - Adds anti-repetition guidelines

## Database Update Required:

Run this SQL to update the balas field:

```sql
ALTER TABLE ai_whatsapp_nodepath 
MODIFY COLUMN balas VARCHAR(255) DEFAULT NULL 
COMMENT 'Timestamp for last response - used for throttling';
```

## Testing the Integration:

1. **Test Onemessage:**
   Send: "I want this section in add response format [onemessage]"
   
2. **Test Stage Detection:**
   Send: "stage: Problem Identification"
   
3. **Test Throttling:**
   Send multiple messages rapidly - should throttle after 4 seconds

## Logging to Watch For:

- `🎯 AI: Detected stage from user input` - Stage detection working
- `⏱️ THROTTLE: Request throttled` - Time throttling active
- `🔗 ONEMESSAGE: Sent combined message` - Onemessage combining working
- `BOT_COMBINED:` in conversation logs - Combined messages being logged