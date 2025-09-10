# Fix for AI Prompt Node Flow Continuation Issue

## Problem Identified
After AI prompt nodes finish processing and send their response, the system terminates the flow instead of checking for and advancing to the next node (often a user_reply node). This breaks the flow continuity.

## Root Cause
The `processAIPromptNode` function in `whatsapp_service.go` is not properly:
1. Checking for next nodes after sending AI response
2. Advancing to user_reply nodes when they exist
3. Setting the waiting_for_reply flag appropriately

## Solution
We need to update the `processAIPromptNode` function to:
1. Always check for next nodes after processing
2. Properly advance to user_reply or waiting_reply_times nodes
3. Continue processing other node types immediately
4. Only terminate flow when there's truly no next node

## Files to Modify
- `internal/whatsapp/whatsapp_service.go` - Update processAIPromptNode function

## Implementation Details

The fix ensures that after an AI prompt node:
1. Check if there's a next node using `flowService.GetNextNode()`
2. If next node is user_reply or waiting_reply_times: advance and set waiting flag
3. If next node is another type: continue processing immediately
4. Only complete flow execution if no next node exists

This maintains proper flow continuity for complex multi-step conversations.
