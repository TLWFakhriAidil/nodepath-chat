# Comprehensive Flow System Fixes

## Issues to Address:
1. Support flows without AI prompt nodes (basic flows)
2. Fix condition node edge selection based on user input
3. Fix condition matching when user input matches edge labels
4. Remove manual and waiting_reply_times nodes from UI
5. Fix duplicate BOT messages in conv_last

## Files to Modify:
- `internal/whatsapp/whatsapp_service.go` - Fix condition processing
- `internal/services/flow_service.go` - Fix condition evaluation
- `src/components/ChatbotBuilder.tsx` - Remove unwanted node types
- `internal/services/ai_whatsapp_service.go` - Fix conversation history duplicates
